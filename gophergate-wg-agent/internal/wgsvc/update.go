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

func UpdatePeer(ctx context.Context, req UpdatePeerRequest) (UpdatePeerResponse, error) {
	if req.Iface == "" {
		return UpdatePeerResponse{}, fmt.Errorf("iface is required")
	}
	if req.PublicKey == "" {
		return UpdatePeerResponse{}, fmt.Errorf("public key is required")
	}
	if len(req.SetAllowedCIDRs) > 0 && len(req.AppendAllowedCIDRs) > 0 {
		return UpdatePeerResponse{}, fmt.Errorf("provide either SetAllowedCIDRs or AppendAllowedCIDRs, not both")
	}

	cli, err := wgctrl.New()
	if err != nil {
		return UpdatePeerResponse{}, fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	pub, err := wgtypes.ParseKey(req.PublicKey)
	if err != nil {
		return UpdatePeerResponse{}, fmt.Errorf("parse public key: %w", err)
	}

	var (
		allowed        []net.IPNet
		replaceAllowed bool
	)
	if len(req.SetAllowedCIDRs) > 0 || len(req.AppendAllowedCIDRs) > 0 {
		var list []string
		if len(req.SetAllowedCIDRs) > 0 {
			list = req.SetAllowedCIDRs
			replaceAllowed = true
		} else {
			list = req.AppendAllowedCIDRs
			replaceAllowed = false
		}
		for _, cidr := range list {
			_, ipn, err := net.ParseCIDR(strings.TrimSpace(cidr))
			if err != nil {
				return UpdatePeerResponse{}, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
			}
			allowed = append(allowed, *ipn)
		}
	}

	var ep *net.UDPAddr
	setEndpoint := false
	if req.Endpoint != "" {
		addr, err := net.ResolveUDPAddr("udp", req.Endpoint)
		if err != nil {
			return UpdatePeerResponse{}, fmt.Errorf("invalid endpoint %q: %w", req.Endpoint, err)
		}
		ep = addr
		setEndpoint = true
	}

	var keepalive *time.Duration
	setKA := false
	if req.KeepaliveSeconds != nil {
		d := time.Duration(*req.KeepaliveSeconds) * time.Second
		keepalive = &d
		setKA = true
	}

	pc := wgtypes.PeerConfig{
		PublicKey: pub,
	}
	if len(allowed) > 0 || replaceAllowed {
		pc.ReplaceAllowedIPs = replaceAllowed
		pc.AllowedIPs = allowed
	}
	if setEndpoint {
		pc.Endpoint = ep
	}
	if setKA {
		pc.PersistentKeepaliveInterval = keepalive
	}

	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{pc}}
	if err := cli.ConfigureDevice(req.Iface, cfg); err != nil {
		return UpdatePeerResponse{}, fmt.Errorf("configure device: %w", err)
	}

	var resp UpdatePeerResponse
	resp.Iface = req.Iface
	resp.PublicKey = pub.String()
	resp.Changed.AllowedIPs = len(allowed) > 0
	resp.Changed.Endpoint = setEndpoint
	resp.Changed.Keepalive = setKA
	return resp, nil
}
