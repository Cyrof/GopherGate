package service

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type TrafficSnapshotter struct {
	trafficSvc    *TrafficService
	iface         string
	interval      time.Duration
	retentionDays int
	log           *zap.SugaredLogger
}

func NewTrafficSnapshotter(
	trafficSvc *TrafficService,
	iface string,
	interval time.Duration,
	retentionDays int,
	log *zap.SugaredLogger,
) *TrafficSnapshotter {
	return &TrafficSnapshotter{
		trafficSvc:    trafficSvc,
		iface:         iface,
		interval:      interval,
		retentionDays: retentionDays,
		log:           log,
	}
}

func (s *TrafficSnapshotter) Run(ctx context.Context) {
	if s.trafficSvc == nil {
		if s.log != nil {
			s.log.Warn("traffic snapshotter disabled: traffic service is nil")
		}
		return
	}
	if s.iface == "" {
		s.iface = "wg0"
	}
	if s.interval <= 0 {
		s.interval = time.Minute
	}
	if s.retentionDays <= 0 {
		s.retentionDays = 3
	}

	if s.log != nil {
		s.log.Infow(
			"traffic snapshotter started",
			"iface", s.iface,
			"interval", s.interval.String(),
			"retention_days", s.retentionDays,
		)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.capture(ctx)

	for {
		select {
		case <-ctx.Done():
			if s.log != nil {
				s.log.Infow("traffic snapshotter stopped", "reason", ctx.Err())
			}
			return
		case <-ticker.C:
			s.capture(ctx)
		}
	}
}

func (s *TrafficSnapshotter) capture(ctx context.Context) {
	if err := s.trafficSvc.CapturePeerTrafficSnapshot(ctx, s.iface); err != nil {
		if s.log != nil {
			s.log.Errorw("traffic snapshot capture failed", "iface", s.iface, "err", err)
		}
		return
	}

	if err := s.trafficSvc.CleanupOldPeerTraffic(ctx, s.retentionDays); err != nil {
		if s.log != nil {
			s.log.Errorw("traffic retention cleanup failed", "retention_days", s.retentionDays, "err", err)
		}
		return
	}

	if s.log != nil {
		s.log.Debugw("traffic snapshot captured", "iface", s.iface)
	}
}
