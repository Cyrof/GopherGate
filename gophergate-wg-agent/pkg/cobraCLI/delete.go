package cobraCLI

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	delIface  string
	delPubKey string
	delAsJSON bool
	delYes    bool
	delForce  bool
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
		if delPubKey == "" && peerID != "" {
			delPubKey = peerID
		}
		if delPubKey == "" {
			return fmt.Errorf("--pubkey (or --id) is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !delYes {
			fmt.Printf("Are you sure you want to delete peer %s on %s? (y/N): ", delPubKey, delIface)
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			if resp != "y" && resp != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		req := wgsvc.DeletePeerRequest{
			Iface:     delIface,
			PublicKey: delPubKey,
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		resp, err := wgsvc.DeletePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)")
			} else {
				Log.Errorw("delete peer failed", "iface", delIface, "err", err)
			}
			return err
		}

		if delAsJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(resp)
		}
		Log.Infow("peer removed", "iface", resp.Iface, "pubKey", resp.PublicKey, "removed", resp.Removed)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&name, "name", "", "Name of the peer to delete")
	deleteCmd.Flags().StringVar(&peerID, "id", "", "Unique ID of the peer to delete")

	deleteCmd.Flags().StringVarP(&delIface, "iface", "i", "wg0", "WireGuard interface name")
	deleteCmd.Flags().StringVarP(&delPubKey, "pubkey", "p", "", "Peer public key (base64)")
	deleteCmd.Flags().BoolVarP(&delAsJSON, "json", "j", false, "Output JSON response")

	// reserved for future interactive safety
	deleteCmd.Flags().BoolVarP(&delYes, "yes", "y", false, "Confirm deletion without prompting (reserved)")
	deleteCmd.Flags().BoolVar(&delForce, "force", false, "Force deletion (reserved)")
	rootCmd.AddCommand(deleteCmd)
}
