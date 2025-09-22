package wgsvc

import (
	"fmt"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GetPeerByNameOrID(iface, name, peerID string) (*PeerStatus, error) {
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
		if peerID != "" && p.PublicKey.String() == peerID {
			return formatPeer(p), nil
		}
	}
	return nil, fmt.Errorf("peer not found: name=%s, id=%s", name, peerID)
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
	return &PeerStatus{
		PublicKey:  p.PublicKey.String(),
		Endpoint:   fmt.Sprint(p.Endpoint),
		AllowedIPs: ips,
		Handshake:  p.LastHandshakeTime.Format(time.RFC3339),
		RxBytes:    uint64(p.ReceiveBytes),
		TxBytes:    uint64(p.TransmitBytes),
		Keepalive:  ka,
	}
}
