package handlers

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"go.uber.org/zap"
)

type peer struct {
	Name      string
	IP        string
	Keepalive string
	PublicKey string
	Endpoint  string
	RxBytes   uint64
	TxBytes   uint64
	Handshake string
	State     string
}

type peerStat struct {
	Label string
	Value int
	State string
}

type enrollmentItem struct {
	Name        string
	Description string
}

type Peers struct {
	log            *zap.SugaredLogger
	data           []peer
	grpc           *grpcclient.Client
	wgIface        string
	peerIPPoolCIDR string
}

func NewPeers(log *zap.SugaredLogger, grpcClient *grpcclient.Client, wgIface string, peerIPPoolCIDR string) *Peers {
	return &Peers{
		log:            log,
		data:           []peer{},
		grpc:           grpcClient,
		wgIface:        wgIface,
		peerIPPoolCIDR: peerIPPoolCIDR,
	}
}

func peerPageData(title string, rows []peer, errorMsg string) map[string]any {
	rows = normalisePeerRows(rows)
	stats, connectedCount := buildPeerStats(rows)
	enrollmentItems := defaultEnrollmentItems()

	return map[string]any{
		"title":            title,
		"pageTitle":        title,
		"activeNav":        "peers",
		"activePage":       "peers",
		"statusText":       fmt.Sprintf("%d Peers Online", connectedCount),
		"error":            errorMsg,
		"peers":            rows,
		"stats":            stats,
		"pendingApprovals": len(enrollmentItems),
		"enrollmentItems":  enrollmentItems,
	}
}

func defaultEnrollmentItems() []enrollmentItem {
	return []enrollmentItem{
		{
			Name:        "Edge Gateway SG-08",
			Description: "requesting key exchange",
		},
		{
			Name:        "Dev Laptop keith-mbp",
			Description: "awaiting route policy",
		},
	}
}

func normaliseEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)

	switch endpoint {
	case "", "-", "—", "--", "<nil>":
		return ""
	default:
		return endpoint
	}
}

func validateEndpoint(endpoint string) error {
	endpoint = normaliseEndpoint(endpoint)
	if endpoint == "" {
		return nil
	}

	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint")
	}

	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("invalid endpoint")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("invalid endpoint")
	}

	return nil
}

func normalisePeerRows(rows []peer) []peer {
	for i := range rows {
		if strings.TrimSpace(rows[i].Name) == "" {
			rows[i].Name = defaultPeerName(rows[i].PublicKey)
		}

		if normaliseEndpoint(rows[i].Endpoint) == "" {
			rows[i].Endpoint = "—"
		}

		if normaliseAllowedIP(rows[i].IP) == "" {
			rows[i].IP = "—"
		}

		if strings.TrimSpace(rows[i].State) == "" {
			rows[i].State = inferPeerState(rows[i].Endpoint, rows[i].Handshake)
		}
	}

	return rows
}

func validateAllowedCIDR(value string) error {
	value = normaliseAllowedIP(value)
	if value == "" {
		return nil
	}

	if _, err := netip.ParsePrefix(value); err != nil {
		return fmt.Errorf("invalid allowed ip")
	}

	return nil
}

func buildPeerStats(rows []peer) ([]peerStat, int) {
	connected := 0
	intermittent := 0
	offline := 0

	for _, row := range rows {
		switch strings.ToUpper(strings.TrimSpace(row.State)) {
		case "CONNECTED":
			connected++
		case "INTERMITTENT":
			intermittent++
		default:
			offline++
		}
	}

	return []peerStat{
		{Label: "Connected", Value: connected, State: "connected"},
		{Label: "Intermittent", Value: intermittent, State: "intermittent"},
		{Label: "Offline", Value: offline, State: "offline"},
	}, connected
}

func inferPeerState(endpoint string, handshake string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" || endpoint == "<nil>" || endpoint == "—" || endpoint == "--" {
		return "OFFLINE"
	}

	handshake = strings.TrimSpace(handshake)
	if handshake == "" {
		return "OFFLINE"
	}

	lastHandshake, err := time.Parse(time.RFC3339, handshake)
	if err != nil {
		return "INTERMITTENT"
	}

	age := time.Since(lastHandshake)
	if age < 0 {
		age = 0
	}

	switch {
	case age <= 3*time.Minute:
		return "CONNECTED"
	case age <= 30*time.Minute:
		return "INTERMITTENT"
	default:
		return "OFFLINE"
	}
}

func defaultPeerName(publicKey string) string {
	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return "Unnamed Peer"
	}

	if len(publicKey) <= 8 {
		return "Peer " + publicKey
	}

	return "Peer " + publicKey[:8]
}

func normaliseAllowedIP(ip string) string {
	ip = strings.TrimSpace(ip)

	switch ip {
	case "", "-", "—", "--", "<nil>":
		return ""
	default:
		return ip
	}
}

func allowedIPExists(rows []peer, allowedIP string, excludePublicKey string) bool {
	allowedIP = normaliseAllowedIP(allowedIP)
	if allowedIP == "" {
		return false
	}

	targetAddr, err := parsePeerIPInput(allowedIP)
	if err != nil {
		return allowedIPStringExists(rows, allowedIP, excludePublicKey)
	}

	return peerIPInUse(rows, targetAddr, excludePublicKey)
}

func allowedIPStringExists(rows []peer, allowedIP string, excludePublicKey string) bool {
	excludePublicKey = strings.TrimSpace(excludePublicKey)

	for _, row := range rows {
		if excludePublicKey != "" && strings.TrimSpace(row.PublicKey) == excludePublicKey {
			continue
		}

		if normaliseAllowedIP(row.IP) == allowedIP {
			return true
		}
	}

	return false
}

func normalisePeerName(name string) string {
	return strings.TrimSpace(name)
}

func peerNameExists(rows []peer, name string) bool {
	name = strings.ToLower(normalisePeerName(name))
	if name == "" {
		return false
	}

	for _, row := range rows {
		if strings.ToLower(normalisePeerName(row.Name)) == name {
			return true
		}
	}

	return false
}

func parseKeepaliveSeconds(value string) (int32, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	value = strings.TrimSuffix(strings.ToLower(value), "s")

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("invalid keepalive value")
	}

	return int32(seconds), nil
}

func (p *Peers) peerIPPool() (netip.Prefix, error) {
	raw := strings.TrimSpace(p.peerIPPoolCIDR)
	if raw == "" {
		return netip.Prefix{}, errPeerIPPoolMissing
	}

	prefix, err := netip.ParsePrefix(raw)
	if err != nil {
		return netip.Prefix{}, errPeerIPPoolInvalid
	}

	return prefix.Masked(), nil
}

func (p *Peers) normaliseCreateAllowedCIDR(value string, existingRows []peer) (string, error) {
	pool, err := p.peerIPPool()
	if err != nil {
		return "", err
	}

	value = normaliseAllowedIP(value)
	if value == "" {
		return allocatePeerCIDR(pool, existingRows)
	}

	addr, err := parsePeerIPInput(value)
	if err != nil {
		return "", errPeerIPInvalid
	}

	if !pool.Contains(addr) {
		return "", errPeerIPOutOfRange
	}

	if peerIPInUse(existingRows, addr, "") {
		return "", errPeerIPAlreadyUsed
	}

	return hostCIDR(addr), nil
}

func parsePeerIPInput(value string) (netip.Addr, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return netip.Addr{}, errPeerIPInvalid
	}

	if strings.Contains(value, "/") {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return netip.Addr{}, err
		}

		addr := prefix.Addr()

		// For one peer assignment, only allow a single host address:
		// IPv4: /32
		// IPv6: /128
		if prefix.Bits() != addr.BitLen() {
			return netip.Addr{}, errPeerIPInvalid
		}

		return addr, nil
	}

	return netip.ParseAddr(value)
}

func allocatePeerCIDR(pool netip.Prefix, existingRows []peer) (string, error) {
	used := usedPeerIPs(existingRows)
	broadcast, hasBroadcast := ipv4BroadcastAddr(pool)

	checked := 0
	for addr := pool.Addr(); pool.Contains(addr); addr = addr.Next() {
		checked++
		if checked > 65536 {
			return "", errPeerIPPoolFull
		}

		if shouldSkipIPv4NetworkAddress(pool, addr) {
			continue
		}

		if hasBroadcast && addr == broadcast && pool.Bits() < 31 {
			continue
		}

		if _, exists := used[addr]; exists {
			continue
		}

		return hostCIDR(addr), nil
	}

	return "", errPeerIPPoolFull
}

func usedPeerIPs(rows []peer) map[netip.Addr]struct{} {
	used := make(map[netip.Addr]struct{})

	for _, row := range rows {
		value := normaliseAllowedIP(row.IP)
		if value == "" {
			continue
		}

		addr, err := parsePeerIPInput(value)
		if err != nil {
			continue
		}

		used[addr] = struct{}{}
	}

	return used
}

func peerIPInUse(rows []peer, addr netip.Addr, excludePublicKey string) bool {
	excludePublicKey = strings.TrimSpace(excludePublicKey)

	for _, row := range rows {
		if excludePublicKey != "" && strings.TrimSpace(row.PublicKey) == excludePublicKey {
			continue
		}

		value := normaliseAllowedIP(row.IP)
		if value == "" {
			continue
		}

		existingAddr, err := parsePeerIPInput(value)
		if err != nil {
			continue
		}

		if existingAddr == addr {
			return true
		}
	}

	return false
}

func hostCIDR(addr netip.Addr) string {
	return fmt.Sprintf("%s/%d", addr.String(), addr.BitLen())
}

func shouldSkipIPv4NetworkAddress(pool netip.Prefix, addr netip.Addr) bool {
	return addr.Is4() && pool.Addr().Is4() && pool.Bits() < 31 && addr == pool.Addr()
}

func ipv4BroadcastAddr(prefix netip.Prefix) (netip.Addr, bool) {
	prefix = prefix.Masked()
	addr := prefix.Addr()

	if !addr.Is4() {
		return netip.Addr{}, false
	}

	bytes := addr.As4()
	ip := binary.BigEndian.Uint32(bytes[:])

	hostBits := 32 - prefix.Bits()
	if hostBits <= 0 {
		return addr, true
	}

	hostMask := uint32(1<<hostBits) - 1
	broadcast := ip | hostMask

	var out [4]byte
	binary.BigEndian.PutUint32(out[:], broadcast)

	return netip.AddrFrom4(out), true
}
