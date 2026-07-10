package handlers

import (
	"fmt"
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
	log     *zap.SugaredLogger
	data    []peer
	grpc    *grpcclient.Client
	wgIface string
}

func NewPeers(log *zap.SugaredLogger, grpcClient *grpcclient.Client, wgIface string) *Peers {
	return &Peers{
		log:     log,
		data:    []peer{},
		grpc:    grpcClient,
		wgIface: wgIface,
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

func normalisePeerRows(rows []peer) []peer {
	for i := range rows {
		if strings.TrimSpace(rows[i].Name) == "" {
			rows[i].Name = defaultPeerName(rows[i].PublicKey)
		}

		endpoint := strings.TrimSpace(rows[i].Endpoint)
		if endpoint == "" || endpoint == "<nil>" || endpoint == "—" || endpoint == "--" {
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
	case "", "-", "--", "<nil>":
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
