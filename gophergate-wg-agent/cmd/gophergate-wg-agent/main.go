package main

import (
	stdlog "log"
	"os"

	"context"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
	"github.com/Cyrof/GopherGate/gophergate-core/envx"
	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/schema/migration"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/pkg/cobraCLI"
)

var Version = "0.1.0"

func main() {
	// load .env if present
	envx.LoadDotenvIfPresent()

	p := paths.ForApp(paths.AppAgent)
	if err := p.Ensure(); err != nil {
		stdlog.Fatalf("failed to ensure paths: %v", err)
	}

	logger, flush := logx.Init(logx.Default(paths.AppAgent))
	defer flush()

	logger.Infow("agent started", "version", Version)

	// initialise database
	cfg := dbx.Default(paths.AppAgent)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		cfg.DSN = dsn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, closePool, err := dbx.Open(ctx, cfg)
	if err != nil {
		logger.Fatalw("failed to connect database", "err", err)
	}
	defer closePool()

	mcfg := &dbx.MigrateConfig{AdvisoryLockKey: 4242}

	if err := dbx.MigrateFS(ctx, pool, migration.FS, ".", mcfg); err != nil {
		logger.Fatalw("migrate failed", "err", err)
	}

	logger.Infow("Database initialised.")

	cobraCLI.SetLogger(logger)

	cobraCLI.Execute()
}
