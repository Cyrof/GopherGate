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
	AllowedIPs          []net.IP
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
		values ($1, $2, $3, $4, $5, $6)
		returning id::text;
	`
	allowed := make([]string, 0, len(p.AllowedIPs))
	for _, ip := range p.AllowedIPs {
		allowed = append(allowed, ip.String())
	}

	var id string
	err := r.db.QueryRow(ctx, q,
		p.Name,
		p.PublicKey,
		p.IPAddress.String(),
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
		select id::text, name, public_key, ip_address, allowed_ips, endpoint, 
			persistent_keepalive, last_handshake, create_at, updated_at
		from peers
		where public_key = $1
		limit 1;
	`
	var p Peer
	var ipStr string
	var allowed []string
	if err := r.db.QueryRow(ctx, q, pub).Scan(
		&p.ID, &p.Name, &p.PublicKey, &ipStr, &allowed, &p.Endpoint,
		&p.PersistentKeepalive, &p.LastHandshake, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	p.IPAddress = net.ParseIP(ipStr)
	for _, a := range allowed {
		if ip := net.ParseIP(a); ip != nil {
			p.AllowedIPs = append(p.AllowedIPs, ip)
		}
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
		add(`allowed_ips = $`+fmt.Sprint(len(args)+1)+`::text[]`, in.ReplaceAllowed)
		changeAllowed = true
	case len(in.AppendAllowed) > 0:
		add(`allowed_ips = (
			select array(
				select distinct x from (
					select unnest(coalesce(allowed_ips, '{}')) as x
					union all
					select unest($`+fmt.Sprint(len(args)+1)+`::text[]) as x
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

	q := `update peers set` + strings.Join(setParts, ", ") + `where public_key = $` + fmt.Sprint(len(args)+1) + `;`
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
