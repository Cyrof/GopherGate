package wgsvc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func CreatePeer(ctx context.Context, req CreatePeerRequest) (CreatePeerResponse, error) {
	// validate input
	if req.Iface == "" {
		return CreatePeerResponse{}, fmt.Errorf("iface is required")
	}
	if req.PublicKey == "" {
		return CreatePeerResponse{}, fmt.Errorf("public key is required")
	}
	if len(req.AllowedCIDRs) == 0 {
		return CreatePeerResponse{}, fmt.Errorf("at least one allowed CIDR is required")
	}

	// open wgctrl client
	cli, err := wgctrl.New()
	if err != nil {
		return CreatePeerResponse{}, fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	// parse inputs into wgtypes
	pub, err := wgtypes.ParseKey(req.PublicKey)
	if err != nil {
		return CreatePeerResponse{}, fmt.Errorf("parse public key: %w", err)
	}

	var allowed []net.IPNet
	for _, cidr := range req.AllowedCIDRs {
		_, ipn, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			return CreatePeerResponse{}, fmt.Errorf("invalid allowed CIDR %q: %w", cidr, err)
		}
		allowed = append(allowed, *ipn)
	}

	var ep *net.UDPAddr
	if req.Endpoint != "" {
		addr, err := net.ResolveUDPAddr("udp", req.Endpoint)
		if err != nil {
			return CreatePeerResponse{}, fmt.Errorf("invalid endpoint %q: %w", req.Endpoint, err)
		}
		ep = addr
	}

	var ka *time.Duration
	if req.KeepaliveSeconds > 0 {
		d := time.Duration(req.KeepaliveSeconds) * time.Second
		ka = &d
	}

	// build the peer and apply to device
	pc := wgtypes.PeerConfig{
		PublicKey:                   pub,
		ReplaceAllowedIPs:           req.ReplaceAllowedIPs || true,
		AllowedIPs:                  allowed,
		Endpoint:                    ep,
		PersistentKeepaliveInterval: ka,
	}
	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{pc}}

	// atomic apply, kernel updates the device with this peer configuration
	if err := cli.ConfigureDevice(req.Iface, cfg); err != nil {
		return CreatePeerResponse{}, fmt.Errorf("configure device: %w", err)
	}

	return CreatePeerResponse{
		Iface:         req.Iface,
		PublicKey:     pub.String(),
		ConfigApplied: true,
	}, nil
}
