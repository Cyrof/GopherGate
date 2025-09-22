package cobraCLI

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	createIface      string
	createName       string
	createPubKey     string
	createAllowed    []string
	createEndpoint   string
	createKeepalive  int
	createReplaceIPs bool
	createAsJSON     bool
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if createPubKey == "" {
			return fmt.Errorf("--pubkey is required")
		}
		if len(createAllowed) == 0 {
			return fmt.Errorf("at least one --allowed CIDR is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		req := wgsvc.CreatePeerRequest{
			Iface:             createIface,
			Name:              createName,
			PublicKey:         createPubKey,
			AllowedCIDRs:      createAllowed,
			Endpoint:          createEndpoint,
			KeepaliveSeconds:  createKeepalive,
			ReplaceAllowedIPs: createReplaceIPs,
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		resp, err := wgsvc.CreatePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)")
			} else {
				Log.Errorw("create peer failed", "iface", createIface, "err", err)
			}
			return err
		}

		if createAsJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(resp)
		}

		Log.Infow("peer added", "name", resp.Name, "iface", resp.Iface, "pubKey", resp.PublicKey, "configApplied", resp.ConfigApplied)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createIface, "iface", "i", "wg0", "WireGuard interface name")
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Friendly name for this peer (optional)")
	createCmd.Flags().StringVarP(&createPubKey, "pubkey", "p", "", "Peer public key (base64, required)")
	createCmd.Flags().StringSliceVarP(&createAllowed, "allowed", "a", nil, "Allowed IPs (CIDR). Repeatable (required)")
	createCmd.Flags().IntVarP(&createKeepalive, "keepalive", "k", 0, "Persistent keepalive in second (0 = disabled)")
	createCmd.Flags().BoolVarP(&createReplaceIPs, "replace-ips", "r", true, "Replace existing AllowedIPs for this peer")
	createCmd.Flags().BoolVarP(&createAsJSON, "json", "j", false, "Output JSON response")

	rootCmd.AddCommand(createCmd)
}
