package wgsvc

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
	PublicKey         string
	AllowedCIDRs      []string
	Endpoint          string
	KeepaliveSeconds  int
	ReplaceAllowedIPs bool
}

type CreatePeerResponse struct {
	Iface         string `json:"iface"`
	PublicKey     string `json:"public_key"`
	ConfigApplied bool   `json:"config_applied"`
}