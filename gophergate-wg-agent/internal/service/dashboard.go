package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
)

type DashboardService struct {
	repo  *data.Repository
	nowFn func() time.Time
}

func NewDashboardService(repo *data.Repository) *DashboardService {
	return &DashboardService{
		repo:  repo,
		nowFn: time.Now,
	}
}

type DashboardResult struct {
	Summary DashboardSummary
	Peers   []DashboardPeer
	Alerts  []DashboardAlert
	Notes   []DashboardNote
}

type DashboardSummary struct {
	PeersOnline          int
	PeersTotal           int
	PeersChangedLastHour int
	TotalRXBytes         uint64
	TotalTXBytes         uint64
	TotalTrafficBytes    uint64
	LastSync             string
	AgentStatus          string
	DegradedNodes        int
}

type DashboardPeer struct {
	PublicKey     string
	Name          string
	Endpoint      string
	AllowedIPs    []string
	RXBytes       uint64
	TXBytes       uint64
	LastHandshake string
	Latency       string
	State         string
	StateReason   string
}

type DashboardAlert struct {
	ID       string
	Title    string
	Body     string
	Severity string
}

type DashboardNote struct {
	Title string
	Body  string
	Level string
}

func (s *DashboardService) Build(ctx context.Context, iface string) (*DashboardResult, error) {
	if iface == "" {
		return nil, fmt.Errorf("iface is required")
	}

	runtimePeers, err := wgsvc.ListPeers(iface)
	if err != nil {
		return nil, fmt.Errorf("list runtime peers: %w", err)
	}

	publicKeys := make([]string, 0, len(runtimePeers))
	for _, p := range runtimePeers {
		publicKeys = append(publicKeys, p.PublicKey)
	}

	nameMap := map[string]string{}
	if s.repo != nil {
		names, err := s.repo.NamesByPublicKeys(ctx, publicKeys)
		if err != nil {
			return nil, fmt.Errorf("load peer names: %w", err)
		}
		nameMap = names
	}

	peers := make([]DashboardPeer, 0, len(runtimePeers))
	for _, rp := range runtimePeers {
		state, reason := derivePeerState(rp.Handshake)

		name := nameMap[rp.PublicKey]
		if name == "" {
			name = shortenKey(rp.PublicKey)
		}

		peers = append(peers, DashboardPeer{
			PublicKey:     rp.PublicKey,
			Name:          name,
			Endpoint:      rp.Endpoint,
			AllowedIPs:    rp.AllowedIPs,
			RXBytes:       rp.RxBytes,
			TXBytes:       rp.TxBytes,
			LastHandshake: rp.Handshake,
			Latency:       "-",
			State:         state,
			StateReason:   reason,
		})
	}

	summary := s.buildSummary(peers)
	alerts := s.buildAlerts(peers, summary)
	notes := s.buildNotes(peers, summary)

	return &DashboardResult{
		Summary: summary,
		Peers:   peers,
		Alerts:  alerts,
		Notes:   notes,
	}, nil
}

func (s *DashboardService) buildSummary(peers []DashboardPeer) DashboardSummary {
	var online int
	var degraded int
	var totalRX uint64
	var totalTX uint64

	for _, p := range peers {
		if p.State == "CONNECTED" {
			online++
		}
		if p.State == "DOWN" || p.State == "INTERMITTENT" {
			degraded++
		}
		totalRX += p.RXBytes
		totalTX += p.TXBytes
	}

	return DashboardSummary{
		PeersOnline:          online,
		PeersTotal:           len(peers),
		PeersChangedLastHour: 0,
		TotalRXBytes:         totalRX,
		TotalTXBytes:         totalTX,
		TotalTrafficBytes:    totalRX + totalTX,
		LastSync:             s.nowFn().UTC().Format(time.RFC3339),
		AgentStatus:          deriveAgentStatus(peers),
		DegradedNodes:        degraded,
	}
}

func (s *DashboardService) buildAlerts(peers []DashboardPeer, summary DashboardSummary) []DashboardAlert {
	alerts := make([]DashboardAlert, 0, 4)

	if summary.PeersTotal == 0 {
		alerts = append(alerts, DashboardAlert{
			ID:       "no-peers",
			Title:    "No peers found",
			Body:     "No peers were found for this interface",
			Severity: "WARNING",
		})
	}

	if summary.PeersTotal > 0 && summary.PeersOnline == 0 {
		alerts = append(alerts, DashboardAlert{
			ID:       "all-peers-down",
			Title:    "No peers online",
			Body:     "All peers appear down or inactive",
			Severity: "CRITICAL",
		})
	}

	for _, p := range peers {
		if p.State == "DOWN" {
			alerts = append(alerts, DashboardAlert{
				ID:       "peer-down-" + p.PublicKey,
				Title:    "Peer down",
				Body:     fmt.Sprintf("Peer %s is down.", p.Name),
				Severity: "WARNING",
			})
		}
	}

	return alerts
}

func (s *DashboardService) buildNotes(peers []DashboardPeer, summary DashboardSummary) []DashboardNote {
	notes := []DashboardNote{
		{
			Title: "Dashboard refreshed",
			Body:  fmt.Sprintf("Dashboard generated for %d peer(s).", summary.PeersTotal),
			Level: "INFO",
		},
	}

	if summary.DegradedNodes > 0 {
		notes = append(notes, DashboardNote{
			Title: "Attention needed",
			Body:  fmt.Sprintf("%d peer(s) require attention", summary.DegradedNodes),
			Level: "WARNING",
		})
	}
	return notes
}

func derivePeerState(handshake string) (string, string) {
	if handshake == "" {
		return "DOWN", "no handshake recorded"
	}

	t, err := time.Parse(time.RFC3339, handshake)
	if err != nil {
		return "INTERMITTENT", "unable to parse handshake timestamp"
	}

	age := time.Since(t)
	switch {
	case age <= 2*time.Minute:
		return "CONNECTED", "recent handshake"
	case age <= 10*time.Minute:
		return "INTERMITTENT", "handshake is getting stale"
	default:
		return "DOWN", "handshake is stale"
	}
}

func deriveAgentStatus(peers []DashboardPeer) string {
	if len(peers) == 0 {
		return "DEGRADED"
	}

	for _, p := range peers {
		if p.State == "DOWN" {
			return "DEGRADED"
		}
	}
	return "UP"
}

func shortenKey(publicKey string) string {
	if len(publicKey) <= 12 {
		return publicKey
	}
	return publicKey[:6] + "..." + publicKey[len(publicKey)-6:]
}
