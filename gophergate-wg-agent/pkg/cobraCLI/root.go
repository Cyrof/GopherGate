package cobraCLI

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"
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
	}
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if DB != nil {
			return nil
		}

		ctx := withShutdown(context.Background())
		cfg := dbx.Config{
			App: paths.AppAgent,
			DSN: os.Getenv("DATABASE_URL"),
		}
		pool, cleanup, err := dbx.Open(ctx, cfg)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		DB = pool
		stop = cleanup
		return nil
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if stop != nil {
			stop()
			stop = nil
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
		os.Exit(1)
	}
}
