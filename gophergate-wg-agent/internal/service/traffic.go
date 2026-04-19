package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
)

type TrafficService struct {
	repo *data.Repository
}

func NewTrafficService(repo *data.Repository) *TrafficService {
	return &TrafficService{repo: repo}
}

func (s *TrafficService) CapturePeerTrafficSnapshot(ctx context.Context, iface string) error {
	if s.repo == nil {
		return fmt.Errorf("repository is nil")
	}
	if iface == "" {
		return fmt.Errorf("iface is required")
	}

	peers, err := wgsvc.ListPeers(iface)
	if err != nil {
		return fmt.Errorf("list peers: %w", err)
	}

	now := time.Now().UTC()
	for _, p := range peers {
		err := s.repo.InsertPeerTrafficSnapshot(ctx, data.PeerTrafficSnapshot{
			Iface:      iface,
			PublicKey:  p.PublicKey,
			RXBytes:    p.RxBytes,
			TXBytes:    p.TxBytes,
			TotalBytes: p.RxBytes + p.TxBytes,
			RecordedAt: now,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type PeerTrafficHistoryResult struct {
	PublicKey string
	Name      string
	Points    []data.PeerTrafficPoint
}

func (s *TrafficService) GetPeerTrafficHistory(ctx context.Context, iface, publicKey, rangeName string) (*PeerTrafficHistoryResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is nil")
	}
	if iface == "" {
		return nil, fmt.Errorf("iface is required")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("public_key is required")
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	if rangeName != "" && rangeName != "24h" {
		return nil, fmt.Errorf("unsupported range: %s", rangeName)
	}

	points, err := s.repo.ListPeerTrafficPoints(ctx, iface, publicKey, since)
	if err != nil {
		return nil, err
	}

	name := ""
	if peer, err := s.repo.GetByPublicKey(ctx, publicKey); err == nil && peer != nil {
		name = peer.Name
	}

	return &PeerTrafficHistoryResult{
		PublicKey: publicKey,
		Name:      name,
		Points:    points,
	}, nil
}

func (s *TrafficService) CleanupOldPeerTraffic(ctx context.Context, retentionDays int) error {
	if s.repo == nil {
		return fmt.Errorf("repository is nil")
	}
	if retentionDays <= 0 {
		return fmt.Errorf("retentionDays must be > 0")
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return s.repo.DeletePeerTrafficOlderThan(ctx, cutoff)
}
