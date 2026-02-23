package cobraCLI

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	delIface  string
	delPubKey string
	delName   string
	delYes    bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm", "remove", "del"},
	Short:   "Delete a WireGuard peer (by public key or name)",
	Long:    `Delete a WireGuard peer from the running interface and remove its record from the database.`,
	Example: `
	# Delete by public key (non-interactive)
	gophergate-wg-agent delete --pubkey <base64> --yes

	# Delete by name (resolve via DB)
	gophergate-wg-agent delete --name alice --yes
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if delPubKey == "" && delName == "" {
			return fmt.Errorf("either --pubkey or --name is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		if DB == nil {
			errf(cmd, "Error: database initialised (required for --name lookup)\n")
			Log.Errorw("delete failed: DB not initialised for name lookup", "iface", delIface, "name", delName)
			return errors.New("database not initialised")
		}

		repo := data.NewRepository(DB)

		pubKey, err := wgsvc.ResolvePublicKey(ctx, repo, delName, delPubKey)
		if err != nil {
			errf(cmd, "Error: failed to resolve name %q: %v\n", delName, err)
			Log.Errorw("resolve name failed", "iface", delIface, "name", delName, "err", err)
			return err
		}

		if !delYes {
			fmt.Printf("Delete peer %s on %s? (y/N): ", selectorLabel(delName, pubKey), delIface)
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				out(cmd, "Aborted.\n")
				Log.Infow("delete aborted by user", "iface", delIface, "pubKey", pubKey)
				return nil
			}
		}

		req := wgsvc.DeletePeerRequest{
			Iface:     delIface,
			PublicKey: pubKey,
			Repo:      repo,
		}

		resp, err := wgsvc.DeletePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				errf(cmd, "Error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)\n")
				Log.Errorw("delete permission error", "iface", delIface, "pubKey", pubKey, "err", err)
			} else {
				errf(cmd, "Error: failed to delete peer on %s: %v\n", delIface, err)
				Log.Errorw("delete peer failed", "iface", delIface, "err", err)
			}
			return err
		}

		outf(cmd, "Peer removed on %s: %s\n", resp.Iface, resp.PublicKey)
		Log.Infow("peer removed", "iface", resp.Iface, "pubKey", resp.PublicKey, "removed", resp.Removed)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&delIface, "iface", "i", "wg0", "WireGuard interface name")
	deleteCmd.Flags().StringVarP(&delPubKey, "pubkey", "p", "", "Peer public key (base64)")
	deleteCmd.Flags().StringVarP(&delName, "name", "n", "", "Peer name (look up in DB)")
	deleteCmd.Flags().BoolVarP(&delYes, "yes", "y", false, "Confirm deletion without prompting (reserved)")

	rootCmd.AddCommand(deleteCmd)
}

func selectorLabel(name, pub string) string {
	if name != "" {
		return fmt.Sprintf("%s (%s)", name, pub)
	}
	return pub
}
