package grpcserver

import (
	"context"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			return d
		}
	}
	return def
}

func envDurationOrDefault(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func Start(logger *zap.SugaredLogger, pool *pgxpool.Pool) error {
	listenAddr := envOrDefault("GRPC_ADDR", ":7443")

	list, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Errorw("grpc listen failed", "addr", listenAddr, "err", err)
		return err
	}

	srv := grpc.NewServer()

	var repo *data.Repository
	if pool != nil {
		repo = data.NewRepository(pool)
	} else {
		logger.Warn("grpc started without db: create via grpc will not persist")
	}

	gatewayv1.RegisterWireGuardServiceServer(srv, NewWireGuardService(repo))
	gatewayv2.RegisterWireGuardServiceServer(srv, NewWireGuardServiceV2(repo))

	if repo != nil {
		iface := envOrDefault("WG_IFACE", "wg0")
		snapshotInterval := envDurationOrDefault("TRAFFIC_SNAPSHOT_INTERVAL", time.Minute)
		retentionDays := envIntOrDefault("TRAFFIC_RETENTION_DAYS", 3)

		trafficSvc := service.NewTrafficService(repo)
		snapshotter := service.NewTrafficSnapshotter(
			trafficSvc,
			iface,
			snapshotInterval,
			retentionDays,
			logger,
		)

		go snapshotter.Run(context.Background())
	}

	logger.Infow("grpc server started", "addr", listenAddr)
	return srv.Serve(list)
}
