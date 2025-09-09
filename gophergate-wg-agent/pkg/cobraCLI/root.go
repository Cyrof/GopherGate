package cobraCLI

import (
  "os"

  "github.com/spf13/cobra"
)

var (
  rootCmd = &cobra.Command{
      Use: "gophergate-wg-agent",
      Short: "A ClI agent for managing WireGuard with Cobra and gRPC.",
      Long: `The gophergate-wg-agent provides a simple Cobra-based CLI to manage
WireGuard interfaces and peers. It also exposes a gRPC server to allow
external tools, such as the gophergate-ui, to interact with the WireGuard
service for automation and integration.`,
      PersistentPreRun: func(cmd *cobra.Command, args []string) {
          // skip command if not runnable
          if !cmd.Runnable() {
            return
          }
        },
    }
)

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}
