package wgsvc

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
)

func Status(ctx context.Context, iface string) (DeviceStatus, error) {
	if iface == "" {
		return DeviceStatus{}, fmt.Errorf("iface is required")
	}

	cli, err := wgctrl.New()
	if err != nil {
		return DeviceStatus{}, fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	dev, err := cli.Device(iface)
	if err != nil {
		return DeviceStatus{}, fmt.Errorf("read device %q: %w", iface, err)
	}

	out := DeviceStatus{
		Interface:    dev.Name,
		ListenPort:   dev.ListenPort,
		FirewallMark: dev.FirewallMark,
	}

	if dev.PrivateKey != (dev.PrivateKey) {
		out.PublicKey = dev.PrivateKey.String()
	}

	for _, p := range dev.Peers {
		var ips []string
		for _, ipn := range p.AllowedIPs {
			ips = append(ips, ipn.String())
		}

		ka := ""
		if p.PersistentKeepaliveInterval > 0 {
			ka = p.PersistentKeepaliveInterval.String()
		}

		hs := ""
		if !p.LastHandshakeTime.IsZero() {
			hs = p.LastHandshakeTime.Format(time.RFC3339)
		}

		out.Peers = append(out.Peers, PeerStatus{
			PublicKey:  p.PublicKey.String(),
			Endpoint:   udpAddrToString(p.Endpoint),
			AllowedIPs: ips,
			Handshake:  hs,
			RxBytes:    uint64(p.ReceiveBytes),
			TxBytes:    uint64(p.TransmitBytes),
			Keepalive:  ka,
		})
	}

	return out, nil
}

func udpAddrToString(a *net.UDPAddr) string {
	if a == nil {
		return ""
	}
	return a.String()
}
