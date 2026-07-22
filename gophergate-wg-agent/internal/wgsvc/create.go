package wgsvc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/ippool"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var ErrIPPoolNotConfigured = errors.New("automatic IP assignment requires a configured IP pool")

func CreatePeer(ctx context.Context, req CreatePeerRequest) (CreatePeerResponse, error) {
	// validate input
	if req.Iface == "" {
		return CreatePeerResponse{}, fmt.Errorf("iface is required")
	}
	if req.PublicKey == "" {
		return CreatePeerResponse{}, fmt.Errorf("public key is required")
	}
	if !req.AutoAssignIP && len(req.AllowedCIDRs) == 0 {
		return CreatePeerResponse{}, fmt.Errorf("at least one allowed CIDR is required when automatic IP assignment is disabled")
	}
	if req.AutoAssignIP {
		if req.IPPool == nil {
			return CreatePeerResponse{}, ErrIPPoolNotConfigured
		}
		if req.Repo == nil {
			return CreatePeerResponse{}, fmt.Errorf("automatic IP assignment requires a database repository")
		}
	}

	// parse inputs into wgtypes
	pub, err := wgtypes.ParseKey(req.PublicKey)
	if err != nil {
		return CreatePeerResponse{}, fmt.Errorf("parse public key: %w", err)
	}

	allowed, err := parseAllowedCIDRs(req.AllowedCIDRs)
	if err != nil {
		return CreatePeerResponse{}, nil
	}

	ep, err := parseEndpoint(req.Endpoint)
	if err != nil {
		return CreatePeerResponse{}, err
	}

	ka := keepaliveDuration(req.KeepaliveSeconds)

	peerRecord := data.Peer{
		Name:                req.Name,
		PublicKey:           pub.String(),
		AllowedIPs:          allowed,
		Endpoint:            optionalString(req.Endpoint),
		PersistentKeepalive: optionalI16(req.KeepaliveSeconds),
	}

	var (
		id           string
		assignedIP   string
		assignedCIDR string
	)

	if req.AutoAssignIP {
		allocation, err := req.Repo.InsertAutoAssigned(ctx, peerRecord, req.IPPool)
		if err != nil {
			return CreatePeerResponse{}, fmt.Errorf("allocate peer IP: %w", err)
		}
		id = allocation.ID
		assignedIP = allocation.AssignedIP.String()
		assignedCIDR = allocation.AssignedCIDR
		allowed = allocation.AllowedIPs

		if err := configurePeer(req, pub, allowed, ep, ka); err != nil {
			cleanupErr := cleanupAutoAllocation(ctx, req.Repo, pub.String())
			if cleanupErr != nil {
				return CreatePeerResponse{}, fmt.Errorf("%w; database cleanup also failed: %v", err, cleanupErr)
			}
			return CreatePeerResponse{}, err
		}
	} else {
		primaryIP := pickPrimaryIP(req.AllowedCIDRs)
		if primaryIP == nil {
			return CreatePeerResponse{}, fmt.Errorf("manual assignment requires a host CIDR such as /32 or /128")
		}
		peerRecord.IPAddress = primaryIP

		if err := configurePeer(req, pub, allowed, ep, ka); err != nil {
			return CreatePeerResponse{}, err
		}

		if req.Repo != nil {
			id, err = req.Repo.Insert(ctx, peerRecord)
			if err != nil {
				rollbackErr := removePeer(req.Iface, pub)
				if rollbackErr != nil {
					return CreatePeerResponse{}, fmt.Errorf("insert peer: %w; WireGuard rollback also failed: %v", err, rollbackErr)
				}
				return CreatePeerResponse{}, fmt.Errorf("insert peer: %w", err)
			}
		}
		assignedIP = primaryIP.String()
		for _, cidr := range req.AllowedCIDRs {
			ip, ipNet, parseErr := net.ParseCIDR(strings.TrimSpace(cidr))
			if parseErr == nil && ip.Equal(primaryIP) {
				assignedCIDR = (&net.IPNet{IP: ip, Mask: ipNet.Mask}).String()
				break
			}
		}
	}

	return CreatePeerResponse{
		Iface:         req.Iface,
		Name:          req.Name,
		PublicKey:     pub.String(),
		ConfigApplied: true,
		ID:            id,
		AssignedIP:    assignedIP,
		AssignedCIDR:  assignedCIDR,
	}, nil
}

func configurePeer(req CreatePeerRequest, pub wgtypes.Key, allowed []net.IPNet, ep *net.UDPAddr, ka *time.Duration) error {
	cli, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{{
		PublicKey:                   pub,
		ReplaceAllowedIPs:           req.ReplaceAllowedIPs,
		AllowedIPs:                  allowed,
		Endpoint:                    ep,
		PersistentKeepaliveInterval: ka,
	}}}
	if err := cli.ConfigureDevice(req.Iface, cfg); err != nil {
		return fmt.Errorf("configure device: %w", err)
	}
	return nil
}

func removePeer(iface string, pub wgtypes.Key) error {
	cli, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()
	return cli.ConfigureDevice(iface, wgtypes.Config{Peers: []wgtypes.PeerConfig{{PublicKey: pub, Remove: true}}})
}

func cleanupAutoAllocation(ctx context.Context, repo *data.Repository, publicKey string) error {
	if repo == nil {
		return nil
	}
	_, err := repo.DeleteByPublicKey(ctx, publicKey)
	return err
}

func parseAllowedCIDRs(cidrs []string) ([]net.IPNet, error) {
	allowed := make([]net.IPNet, 0, len(cidrs))
	seen := make(map[string]struct{}, len(cidrs))
	for _, cidr := range cidrs {
		_, ipn, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("invalid allowed CIDR %q: %w", cidr, err)
		}
		canonical := ipn.String()
		if _, exists := seen[canonical]; exists {
			continue
		}
		seen[canonical] = struct{}{}
		allowed = append(allowed, *ipn)
	}
	return allowed, nil
}

func parseEndpoint(endpoint string) (*net.UDPAddr, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, nil
	}
	addr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}
	return addr, nil
}

func keepaliveDuration(seconds int) *time.Duration {
	if seconds <= 0 {
		return nil
	}
	d := time.Duration(seconds) * time.Second
	return &d
}

func IsIPPoolExhausted(err error) bool {
	return errors.Is(err, ippool.ErrExhausted)
}
