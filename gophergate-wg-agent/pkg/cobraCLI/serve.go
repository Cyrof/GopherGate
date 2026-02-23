package cobraCLI

import (
	"errors"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/grpcserver"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if DB == nil {
				return errors.New("database not initialised; cannot start gRPC server")
			}
			return grpcserver.Start(Log, DB)
		},
	}
}

func init() {
	rootCmd.AddCommand(newServeCmd())
}
