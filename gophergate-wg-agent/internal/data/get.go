package data

import (
	"context"
	"net"
	"strings"
)

func (r *Repository) GetPublicKeyByName(ctx context.Context, name string) (string, error) {
	const q = `
		select public_key
		from peers
		where name = $1
		order by updated_at desc
		limit 1;
	`
	var pub string
	if err := r.db.QueryRow(ctx, q, name).Scan(&pub); err != nil {
		return "", err
	}
	return pub, nil
}

func (r *Repository) GetByPublicKey(ctx context.Context, pub string) (*Peer, error) {
	const q = `
		select id::text, name, public_key, ip_address::text, allowed_ips::text[], endpoint, 
			persistent_keepalive, last_handshake, created_at, updated_at
		from peers
		where public_key = $1
		limit 1;
	`
	var p Peer
	var ipStr *string
	var allowed []string
	if err := r.db.QueryRow(ctx, q, pub).Scan(
		&p.ID, &p.Name, &p.PublicKey, &ipStr, &allowed, &p.Endpoint,
		&p.PersistentKeepalive, &p.LastHandshake, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if ipStr != nil {
		p.IPAddress = net.ParseIP(*ipStr)
	}

	for _, a := range allowed {
		_, ipn, err := net.ParseCIDR(a)
		if err != nil {
			ip := net.ParseIP(strings.TrimSpace(a))
			if ip == nil {
				continue
			}
			mask := net.CIDRMask(32, 32)
			if ip.To4() == nil {
				mask = net.CIDRMask(128, 128)
			}
			ipn = &net.IPNet{IP: ip, Mask: mask}
		}
		p.AllowedIPs = append(p.AllowedIPs, *ipn)
	}
	return &p, nil
}

func (r *Repository) NamesByPublicKeys(ctx context.Context, pubs []string) (map[string]string, error) {
	if len(pubs) == 0 {
		return map[string]string{}, nil
	}
	const q = `
		select public_key, name
		from peers
		where public_key = any($1::text[])
	`
	rows, err := r.db.Query(ctx, q, pubs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string, len(pubs))
	for rows.Next() {
		var pk, name string
		if err := rows.Scan(&pk, &name); err != nil {
			return nil, err
		}
		out[pk] = name
	}
	return out, rows.Err()
}
