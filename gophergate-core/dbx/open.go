package dbx

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, func(), error) {
	dsn, err := resolveDSN(cfg)
	if err != nil {
		return nil, func() {}, err
	}

	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, func() {}, fmt.Errorf("parse dsn: %w", err)
	}

	pc.MaxConns = pickI32(cfg.MaxConns, 4)
	pc.MinConns = pickI32(cfg.MinConns, 0)
	pc.MaxConnLifetime = pickDur(cfg.MaxConnLifetime, 30*time.Minute)
	pc.MaxConnIdelTime = pickDur(cfg.MaxConnIdleTime, 5*time.Minute)
	pc.HealthCheckPeriod = pickDur(cfg.HealthCheckFreq, 30*time.Second)

	cctx, cancel := context.WithTimeout(ctx, pickDur(cfg.ConnectTimeout, 5*time.Second))
	defer cancel()

	pool, err := pgxpool.NewWithConfig(cctx, pc)
	if err != nil {
		return nil, func() {}, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(cctx); err != nil {
		pool.Close()
		return nil, func() {}, fmt.Errorf("ping db: %w", err)
	}
	return pool, pool.Close, nil
}

func resolveDSN(cfg Config) (string, error) {
	if cfg.DSN != "" {
		return cfg.DSN, nil
	}
	if env := os.Getenv("DATABASE_URL"); env != "" {
		return env, nil
	}
	host := pickStr(cfg.Host, "127.0.0.1")
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	user := pickStr(cfg.User, "gg_admin")
	db := pickStr(cfg.DBName, "gophergate")
	ssl := pickStr(cfg.SSLMode, "disabled")
	pass := cfg.Password

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, pass, host, port, db, ssl), nil
}
