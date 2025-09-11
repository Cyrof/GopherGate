package cobraCLI

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	newIP   string
	newName string
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"update"},
	Short:   "Edit details of an existing WireGuard peer (placeholder, no backend yet)",
	Long: `The edit command will be used to update details of an existing WireGuard
peer, such as changing its name or assigned IP address. At present, this
command is only a placeholder - the WireGuard service integration is not
yet implemented`,
	Example: `
	# Edit a peer's name
	gophergate-wg-agent edit --id 123 --name alice-renamed

	# Edit a peer's IP address
	gophergate-wg-agent edit --name bob --ip 10.0.0.42

	# Edit both name and IP address
	gophergate-wg-agent edit --id 456 --name carol --ip 10.0.0.99
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if name == "" && peerID == "" {
			return errors.New("you must specify either --name or --id to edit a peer")
		}
		if name != "" && peerID != "" {
			return errors.New("please specify only one of --name or --id, not both")
		}

		// make sure name or ip is specified
		if newName == "" && newIP == "" {
			return errors.New("you must specify at least one field to update (--name or --ip)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("edit command is not yet implemented")
	},
}

func init() {
	editCmd.Flags().StringVar(&name, "name", "", "Name of the peer to edit")
	editCmd.Flags().StringVar(&peerID, "id", "", "Unique ID of the peer to edit")
	editCmd.Flags().StringVar(&newName, "new-name", "", "New name for the peer")
	editCmd.Flags().StringVar(&newIP, "ip", "", "New IP address for the peer")

	rootCmd.AddCommand(editCmd)
}
