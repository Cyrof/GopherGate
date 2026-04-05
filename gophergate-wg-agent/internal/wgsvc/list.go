package wgsvc

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
)

func ListPeers(iface string) ([]PeerStatus, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface is required")
	}

	cli, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	dev, err := cli.Device(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to read device %s: %w", iface, err)
	}

	var peers []PeerStatus
	for _, p := range dev.Peers {
		peers = append(peers, *formatPeer(p))
	}
	return peers, nil
}
