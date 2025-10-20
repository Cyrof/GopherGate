package data

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Peer struct {
	ID                  string
	Name                string
	PublicKey           string
	IPAddress           net.IP
	AllowedIPs          []net.IPNet
	Endpoint            *string
	PersistentKeepalive *int16
	LastHandshake       *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

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
	err := r.db.QueryRow(ctx, q,
		p.Name,
		p.PublicKey,
		ipStr,
		allowed,
		p.Endpoint,
		p.PersistentKeepalive,
	).Scan(&id)
	return id, err
}

func (r *Repository) DeleteByPublicKey(ctx context.Context, publicKey string) (string, error) {
	const q = `
		delete from peers
		where public_key = $1
		returning id::text;
	`
	var id string
	if err := r.db.QueryRow(ctx, q, publicKey).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

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

type UpdatePeerDBInput struct {
	ReplaceAllowed []string
	AppendAllowed  []string

	SetEndpoint bool
	Endpoint    *string

	SetKeepalive bool
	Keepalive    *int16
}

type UpdatePeerDBResult struct {
	ChangeAllowed   bool
	ChangeEndpoint  bool
	ChangeKeepalive bool
}

func (r *Repository) UpdateByPublicKey(ctx context.Context, pubkey string, in UpdatePeerDBInput) (UpdatePeerDBResult, error) {
	var (
		setParts []string
		args     []any
	)
	add := func(expr string, v any) {
		setParts = append(setParts, expr)
		args = append(args, v)
	}

	// allowed IPs
	var changeAllowed bool
	switch {
	case len(in.ReplaceAllowed) > 0:
		add(`allowed_ips = $`+fmt.Sprint(len(args)+1)+`::net[]`, in.ReplaceAllowed)
		changeAllowed = true

	case len(in.AppendAllowed) > 0:
		add(`allowed_ips = (
			select array(
				select distinct x from (
				select unnest(coalesce(allowed_ips, '{}::inet[]')) as x
					union all
					select unest($`+fmt.Sprint(len(args)+1)+`::inet[]) as x
				) t
			)
		)`, in.AppendAllowed)
		changeAllowed = true
	}

	// endpoint
	changeEndpoint := false
	if in.SetEndpoint {
		changeEndpoint = true
		if in.Endpoint == nil || strings.TrimSpace(*in.Endpoint) == "" {
			setParts = append(setParts, `endpoint = NULL`)
		} else {
			add(`endpoint = $`+fmt.Sprint(len(args)+1), *in.Endpoint)
		}
	}

	// keepalive
	changeKeepalive := false
	if in.SetKeepalive {
		changeKeepalive = true
		if in.Keepalive == nil || *in.Keepalive <= 0 {
			setParts = append(setParts, `persistent_keepalive = NULL`)
		} else {
			add(`persistent_keepalive = $`+fmt.Sprint(len(args)+1), *in.Keepalive)
		}
	}

	// nothing to do?
	if len(setParts) == 0 {
		return UpdatePeerDBResult{}, nil
	}

	// always bump updated_at
	setParts = append(setParts, `updated_at = NOW()`)

	q := `update peers set ` + strings.Join(setParts, ", ") + ` where public_key = $` + fmt.Sprint(len(args)+1) + `;`
	args = append(args, pubkey)

	ct, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return UpdatePeerDBResult{}, err
	}
	if ct.RowsAffected() == 0 {
		return UpdatePeerDBResult{}, fmt.Errorf("no row matched public key")
	}

	return UpdatePeerDBResult{
		ChangeAllowed:   changeAllowed,
		ChangeEndpoint:  changeEndpoint,
		ChangeKeepalive: changeKeepalive,
	}, nil
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

type BootstrapPeer struct {
	PublicKey string
	Allowed   []string
	Endpoint  *string
	Keepalive *int16
	Name      string
}

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
