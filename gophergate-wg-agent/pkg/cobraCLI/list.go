package cobraCLI

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all WireGuard peers (placeholder, no backend yet)",
	Long: `The list commmand will be used to retrieve and display all configured
WireGuard peers, including their names, IDs, and assigned IP addresses.
At present, this command is only a placeholder - the WireGuard service
integration is not yet implemented.`,
	Example: `
	# List all peers
	gophergate-wg-agent list

	# List all peers in JSON format (future flag support)
	gophergate-wg-agent list --output json
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("list command is not yet implemented")
	},
}

func init() {
	listCmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	rootCmd.AddCommand(listCmd)
}
