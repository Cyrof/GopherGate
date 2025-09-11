package cobraCLI

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	yes   bool
	force bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm", "remove", "del"},
	Short:   "Delete an existing WireGuard peer (placeholder, no backend yet)",
	Long: `The delete command will remove a WireGuard peer and its associated
configuration. At present, this command is only a placeholder - the
WireGuard service integration is not yet implemented`,
	Example: `
	# Delete by name (will prompt for confirmation in the future)
	gophergate-wg-agent delete --name alice

	# Delete by ID with non-interactive confirmation
	gophergate-wg-agent delete --id 123 --yes

	# Force delete (future behavior may skip safety checks)
	gophergate-wg-agent delete --name bob --force --yes
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if name == "" && peerID == "" {
			return errors.New("you must specify either --name or --id")
		}
		if name != "" && peerID != "" {
			return errors.New("please specify only one of --name or --id, not both")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("delete command is not yet implemented")
	},
}

func init() {
	deleteCmd.Flags().StringVar(&name, "name", "", "Name of the peer to delete")
	deleteCmd.Flags().StringVar(&peerID, "id", "", "Unique ID of the peer to delete")
	deleteCmd.Flags().BoolVarP(&yes, "yes", "y", false, "Confirm deletion without prompting (future use)")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Force deletion (future use)")

	rootCmd.AddCommand(deleteCmd)
}
