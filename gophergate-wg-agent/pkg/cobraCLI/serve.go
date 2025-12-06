package cobraCLI

import (
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/grpcserver"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpcserver.Start(Log)
		},
	}
}

func init() {
	rootCmd.AddCommand(newServeCmd())
}
