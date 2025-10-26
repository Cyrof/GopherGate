package wgsvc

import (
	"context"
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func DeletePeer(ctx context.Context, req DeletePeerRequest) (DeletePeerResponse, error) {
	if req.Iface == "" {
		return DeletePeerResponse{}, fmt.Errorf("iface is required")
	}
	if req.PublicKey == "" {
		return DeletePeerResponse{}, fmt.Errorf("public key is required")
	}

	cli, err := wgctrl.New()
	if err != nil {
		return DeletePeerResponse{}, fmt.Errorf("wgctrl new: %w", err)
	}
	defer cli.Close()

	pub, err := wgtypes.ParseKey(req.PublicKey)
	if err != nil {
		return DeletePeerResponse{}, fmt.Errorf("parse public key: %w", err)
	}

	pc := wgtypes.PeerConfig{
		PublicKey: pub,
		Remove:    true,
	}
	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{pc}}

	if err := cli.ConfigureDevice(req.Iface, cfg); err != nil {
		return DeletePeerResponse{}, fmt.Errorf("configure device: %w", err)
	}

	return DeletePeerResponse{
		Iface:     req.Iface,
		PublicKey: pub.String(),
		Removed:   true,
	}, nil
}
