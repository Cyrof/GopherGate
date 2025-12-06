package grpcserver

import (
	"net"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Start(logger *zap.SugaredLogger) error {
	listenAddr := envOrDefault("GRPC_ADDR", ":7443")

	list, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Errorw("grpc listen failed", "addr", listenAddr, "err", err)
		return err
	}

	srv := grpc.NewServer()

	gatewayv1.RegisterWireGuardServiceServer(srv, NewWireGuardService())

	logger.Infow("grpc server started", "addr", listenAddr)
	return srv.Serve(list)
}
