package main

import (
	"context"
	stdlog "log"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
	"github.com/Cyrof/GopherGate/gophergate-core/envx"
	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/auth"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/web"
	"github.com/Cyrof/GopherGate/gophergate-ui/migrations"
)

var Version = "0.1.0"

func main() {
	envx.LoadDotenvIfPresent()

	p := paths.ForApp(paths.AppUI)
	if err := p.Ensure(); err != nil {
		stdlog.Fatalf("failed to ensure paths: %v", err)
	}

	logger, flush := logx.Init(logx.Default(paths.AppUI))
	defer flush()

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatalw("config error", "err", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, closePool, err := dbx.Open(ctx, cfg.DB)
	if err != nil {
		logger.Fatalw("database connection failed", "err", err)
	}
	defer closePool()
	logger.Infow("database connected")

	// run migrations from embedded FS
	mcfg := &dbx.MigrateConfig{AdvisoryLockKey: 4243}
	if err := dbx.MigrateFS(ctx, pool, migrations.FS, ".", mcfg); err != nil {
		logger.Fatalw("migration failed", "err", err)
	}
	logger.Infow("migrations applied")

	// seed default admin user (only run once if no admin exists)
	if err := auth.SeedDefaultAdmin(ctx, pool, logger); err != nil {
		logger.Fatalw("seed admin failed", "err", err)
	}

	logger.Infow("ui starting", "version", Version, "http", cfg.HTTPAddr, "grpc", cfg.GRPCAddr, "env", cfg.Env)

	if err := web.Run(cfg, logger, pool); err != nil {
		logger.Fatalw("http server error", "err", err)
	}
}
