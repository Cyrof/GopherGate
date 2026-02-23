package grpcserver

import (
	"net"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
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

	logger.Infow("grpc server started", "addr", listenAddr)
	return srv.Serve(list)
}
