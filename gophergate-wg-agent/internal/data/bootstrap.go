package data

import (
	"context"
	"net"
	"strings"
)

func (r *Repository) BootstrapPeers(ctx context.Context) ([]BootstrapPeer, error) {
	const q = `
		select name, public_key, allowed_ips::text[], endpoint, persistent_keepalive
		from peers
		order by created_at asc;
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BootstrapPeer
	for rows.Next() {
		var name, pub string
		var allowedRaw []string
		var endpoint *string
		var ka *int16

		if err := rows.Scan(&name, &pub, &allowedRaw, &endpoint, &ka); err != nil {
			return nil, err
		}

		allowedCIDRs := make([]string, 0, len(allowedRaw))
		for _, s := range allowedRaw {
			ss := strings.TrimSpace(s)
			if ss == "" {
				continue
			}
			if _, _, err := net.ParseCIDR(ss); err == nil {
				allowedCIDRs = append(allowedCIDRs, ss)
				continue
			}

			if ip := net.ParseIP(ss); ip != nil {
				if ip.To4() != nil {
					allowedCIDRs = append(allowedCIDRs, ip.String()+"/32")
				} else {
					allowedCIDRs = append(allowedCIDRs, ip.String()+"/128")
				}
			}
		}

		out = append(out, BootstrapPeer{
			PublicKey: pub,
			Allowed:   allowedCIDRs,
			Endpoint:  endpoint,
			Keepalive: ka,
			Name:      name,
		})
	}
	return out, rows.Err()
}
