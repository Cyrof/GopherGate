package data

import (
	"context"
	"fmt"
)

func (r *Repository) Insert(ctx context.Context, p Peer) (string, error) {
	const q = `
		insert into peers (name, public_key, ip_address, allowed_ips, endpoint, persistent_keepalive)	
		values ($1, $2, $3, $4::inet[], $5, $6)
		returning id::text;
	`
	allowed := make([]string, 0, len(p.AllowedIPs))
	for _, ipn := range p.AllowedIPs {
		allowed = append(allowed, ipn.String())
	}

	var ipStr *string
	if p.IPAddress != nil {
		s := p.IPAddress.String()
		ipStr = &s
	}

	var id string
	if err := r.db.QueryRow(ctx, q,
		p.Name,
		p.PublicKey,
		ipStr,
		allowed,
		p.Endpoint,
		p.PersistentKeepalive,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("insert peer: %w", err)
	}
	return id, nil
}
