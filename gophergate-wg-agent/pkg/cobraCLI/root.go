package cobraCLI

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var (
	DB      *pgxpool.Pool
	stop    func()
	rootCmd = &cobra.Command{
		Use:   "gophergate-wg-agent",
		Short: "A ClI agent for managing WireGuard with Cobra and gRPC.",
		Long: `The gophergate-wg-agent provides a simple Cobra-based CLI to manage
WireGuard interfaces and peers. It also exposes a gRPC server to allow
external tools, such as the gophergate-ui, to interact with the WireGuard
service for automation and integration.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if DB != nil {
			return nil
		}

		ctx := withShutdown(context.Background())

		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			errf(cmd, "Warning: DATABASE_URL is not set; DB-backend features may be limited.\n")
			Log.Warn("DATABASE_URL not set; continuing without DB")
			return nil
		}

		pool, cleanup, err := dbx.Open(ctx, dbx.Config{DSN: dsn})
		if err != nil {
			errf(cmd, "Error: failed to open database: %v\n", err)
			Log.Errorw("open db failed", "err", err)
			return fmt.Errorf("open db: %w", err)
		}

		DB = pool
		stop = cleanup
		Log.Infow("database initialised")
		return nil
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if stop != nil {
			stop()
			stop = nil
			Log.Debug("database connection closed")
		}
	}
}

func withShutdown(ctx context.Context) context.Context {
	c, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c.Done()
		cancel()
	}()
	return c
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Log.Errorw("unexpected error occured", "err", err)
		os.Exit(1)
	}
}
