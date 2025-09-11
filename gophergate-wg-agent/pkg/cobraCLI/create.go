package cobraCLI

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "Create a new WireGuard peer (placholder, no backend yet)",
	Long: `The create command will be used to provide a new WireGuard peer
for a user, including generating keys and configuration. At present, this
command is only a placeholder - the WireGuard service integration is not
yet implemented, so running it will not create any peers.`,
	Example: `
  # Create a new peer with default settings
  gophergate-wg-agent create

  # Create a peer with a specific name
  gophergate-wg-agent create --name alice

  # Create a peer and specify an IP address
  gophergate-wg-agent create --name bob --ip 10.0.0.2
  `,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("create command is not yet implemented")
	},
}

func init() {
	createCmd.Flags().StringVar(&name, "name", "", "Name of user")
	createCmd.Flags().StringVar(&ip, "ip", "", "IP address of the user connection")

	rootCmd.AddCommand(createCmd)
}
