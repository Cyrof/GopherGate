package data

import (
	"context"
	"net"
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
