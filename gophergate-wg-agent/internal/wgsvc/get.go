package wgsvc

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GetPeerByPublicKey(iface, pubkey string) (*PeerStatus, error) {
	ctx, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("wgctrl init error: %w", err)
	}
	defer ctx.Close()

	dev, err := ctx.Device(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to read device %s: %w", iface, err)
	}
	for _, p := range dev.Peers {
		if p.PublicKey.String() == pubkey {
			return formatPeer(p), nil
		}
	}
	return nil, fmt.Errorf("peer not found on %s", iface)
}

func formatPeer(p wgtypes.Peer) *PeerStatus {
	ips := []string{}
	for _, ipn := range p.AllowedIPs {
		ips = append(ips, ipn.String())
	}
	ka := ""
	if p.PersistentKeepaliveInterval > 0 {
		ka = p.PersistentKeepaliveInterval.String()
	}
	hs := ""
	if !p.LastHandshakeTime.IsZero() {
		hs = p.LastHandshakeTime.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return &PeerStatus{
		PublicKey:  p.PublicKey.String(),
		Endpoint:   fmt.Sprint(p.Endpoint),
		AllowedIPs: ips,
		Handshake:  hs,
		RxBytes:    uint64(p.ReceiveBytes),
		TxBytes:    uint64(p.TransmitBytes),
		Keepalive:  ka,
	}
}
