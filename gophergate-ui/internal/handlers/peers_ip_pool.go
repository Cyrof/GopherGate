package handlers

import (
	"context"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
)

func (p *Peers) loadIPPoolStatus(ctx context.Context) ipPoolStatus {
	resp, err := p.grpc.GetIPPool(ctx, &gatewayv2.GetIPPoolRequest{
		Iface: p.wgIface,
	})
	if err != nil {
		p.log.Warnw("peer.ip_pool.grpc_error", "iface", p.wgIface, "err", err)
		return ipPoolStatus{}
	}

	if resp == nil || resp.Pool == nil {
		p.log.Warn("peer.ip_pool.empty_response", "iface", p.wgIface)
		return ipPoolStatus{}
	}

	pool := resp.Pool
	return ipPoolStatus{
		Enabled: true,
		HasAvailable: pool.Available > 0 && pool.NextAvailableIp != "",
		RangeStart: pool.RangeStart,
		RangeEnd: pool.RangeEnd,
		Capacity: pool.Capacity,
		Allocated: pool.Allocated,
		Available: pool.Available,
		NextAvailableIP: pool.NextAvailableIp,
		ReservedIPs: append([]string(nil), pool.ReservedIps...),
	}
}