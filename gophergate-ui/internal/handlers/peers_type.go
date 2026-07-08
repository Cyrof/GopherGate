package handlers

import (
	"fmt"
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

type enrollmentItems struct {
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
	rows = normalisePeerRow(rows)
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

func defaultEnrollmentItems() []enrollmentItems {
	return []enrollmentItems{
		{
			Name:        "Edge Gateway SG-08",
			Description: "requesting key exchange",
		},
		{
			Name:        "Dev Laptop cyrof-mbp",
			Description: "awaiting route policy",
		},
	}
}

func normalisePeerRow(rows []peer) []peer {
	for i := range rows {
		if strings.TrimSpace(rows[i].Name) == "" {
			rows[i].Name = fmt.Sprintf("Peer #%02d", i+1)
		}
		if strings.TrimSpace(rows[i].Endpoint) == "" {
			rows[i].Endpoint = "-"
		}
		if strings.TrimSpace(rows[i].IP) == "" {
			rows[i].IP = "-"
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
	if endpoint == "" || endpoint == "-" || endpoint == "--" {
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
