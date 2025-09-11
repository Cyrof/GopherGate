package cobraCLI

import (
	"errors"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g", "retrieve"},
	Short:   "Retrieve details of a WireGuard peer (placeholder, no backend yet)",
	Long: `The get command will be used to fetch details of a WireGuard peer,
such as public key, IP address, and connection status. At present, this
command is only a placeholder - the WireGuard service integration is not
not yet implemented, so running it will not return any peer information`,
	Example: `
	# Retrieve details for a specific peer by name
	gophergate-wg-agent get --name alice

	# Retrieve details for a peer by ID
	gophergate-wg-agent get --id 123
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// make sure either name or id is provided
		if name == "" && peerID == "" {
			return errors.New("you must specify either --name or --id")
		}
		if name != "" && peerID != "" {
			return errors.New("please specify only one of --name or --id, not both")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("get command is not yet implemented")
	},
}

func init() {
	getCmd.Flags().StringVar(&name, "name", "", "Name of the user peer")
	getCmd.Flags().StringVar(&peerID, "id", "", "Unique ID of the peer")

	rootCmd.AddCommand(getCmd)
}
