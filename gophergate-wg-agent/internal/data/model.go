package data

import (
	"net"
	"time"
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

type PeerTrafficSnapshot struct {
	ID         string
	Iface      string
	PublicKey  string
	RXBytes    uint64
	TXBytes    uint64
	TotalBytes uint64
	RecordedAt time.Time
}

type PeerTrafficPoint struct {
	Timestamp  time.Time
	RXBytes    uint64
	TXBytes    uint64
	TotalBytes uint64
}
