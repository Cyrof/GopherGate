package main

import (
	"fmt"
	stdlog "log"

	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/httpserver"
)

var Version = "0.1.0"

func main() {
	// In the main function, it should run the configuration setup from the `config` folder
	// and also the initialisation for the httpserver that is setup using gin from the `httpserver` folder
	fmt.Println("Ui skeleton running...")

	p := paths.ForApp("gophergate-ui")
	if err := p.Ensure(); err != nil {
		stdlog.Fatalf("failed to ensure paths: %v", err)
	}

	logger, flush := logx.Init(logx.Default("gophergate-ui"))
	defer flush()

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatalw("config error", "err", err)
	}

	logger.Infow("ui starting", "version", Version, "http", cfg.HTTPAddr, "grpc", cfg.GRPCAddr, "env", cfg.Env)

	if err := httpserver.Run(cfg, logger); err != nil {
		logger.Fatalw("http server error", "err", err)
	}
}
