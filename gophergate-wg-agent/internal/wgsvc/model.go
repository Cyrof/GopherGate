package wgsvc

import "github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"

type DeviceStatus struct {
	Interface    string       `json:"interface"`
	ListenPort   int          `json:"list_port"`
	FirewallMark int          `json:"firewall_mark"`
	PublicKey    string       `json:"public_key"`
	Peers        []PeerStatus `json:"peers"`
}

type PeerStatus struct {
	PublicKey  string   `json:"public_key"`
	Endpoint   string   `json:"endpoint"`
	AllowedIPs []string `json:"allowed_ips"`
	Handshake  string   `json:"latest_handshake"`
	RxBytes    uint64   `json:"rx_bytes"`
	TxBytes    uint64   `json:"tx_bytes"`
	Keepalive  string   `json:"keepalive"`
}

type CreatePeerRequest struct {
	Iface             string
	Name              string
	PublicKey         string
	AllowedCIDRs      []string
	Endpoint          string
	KeepaliveSeconds  int
	ReplaceAllowedIPs bool
	Repo              *data.Repository
}

type CreatePeerResponse struct {
	Iface         string `json:"iface"`
	Name          string `json:"name,omitempty"`
	PublicKey     string `json:"public_key"`
	ConfigApplied bool   `json:"config_applied"`
	ID            string
}

type UpdatePeerRequest struct {
	Iface              string
	PublicKey          string
	SetAllowedCIDRs    []string
	AppendAllowedCIDRs []string
	Endpoint           string
	KeepaliveSeconds   *int
	Repo               *data.Repository
}

type UpdatePeerResponse struct {
	Iface     string `json:"iface"`
	PublicKey string `json:"public_key"`

	Changed struct {
		AllowedIPs bool `json:"allowed_ips"`
		Endpoint   bool `json:"endpoint"`
		Keepalive  bool `json:"keepalive"`
	} `json:"changed"`
}

type DeletePeerRequest struct {
	Iface     string
	PublicKey string
	Repo      *data.Repository
}

type DeletePeerResponse struct {
	Iface     string `json:"iface"`
	PublicKey string `json:"public_key"`
	Removed   bool   `json:"removed"`
}
