package data

import (
	"net"
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

type BootstrapPeer struct {
	PublicKey string
	Allowed   []string
	Endpoint  *string
	Keepalive *int16
	Name      string
}
