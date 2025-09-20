package cobraCLI

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	statusIface  string
	statusAsJson bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show WireGuard dsice and peer status",
	Long: `Show current WireGuard dsice information (listen port, public key, firewall mark)
and all peers (endpoint, allowed IPs, last handshake, bytes, keepalive).

In ds, this reads the dsice create by the linuxserver/wireguard container (host networking)
using wgctrl-go (generic netlink). No 'wg' CLI is required.`,
	Example: `
	# Show status of wg0 (human-readable logs)
	gophergate-wg-agent status --iface wg0
	
	# Same with shorthand flags
	gophergate-wg-agent status -i wg0
	
	# JSON output for piping to tools (e.g., jp)
	gophergate-wg-agent status -i wg0 -j | jq .
	
	# Only peers in JSON, pretty-printed
	gophergate-wg-agent status -i wg0 -j | jq '.peers[]'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		ds, err := wgsvc.Status(ctx, statusIface)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep")
			} else {
				Log.Errorw("status failed", "iface", statusIface, "err", err)
			}
		}

		if statusAsJson {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(ds)
		}

		// human logs
		Log.Infow("wg dsice",
			"name", ds.Interface,
			"listenPort", ds.ListenPort,
			"fwMark", ds.FirewallMark,
			"pubKey", ds.PublicKey,
			"peers", len(ds.Peers),
		)
		for _, p := range ds.Peers {
			Log.Infow("peer",
				"pubKey", p.PublicKey,
				"endpoint", p.Endpoint,
				"allowedIPs", strings.Join(p.AllowedIPs, ", "),
				"latestHandshake", p.Handshake,
				"rxBytes", p.RxBytes,
				"txBytes", p.TxBytes,
				"keepalive", p.Keepalive,
			)
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusIface, "iface", "i", "wg0", "WireGuard interface name")
	statusCmd.Flags().BoolVarP(&statusAsJson, "json", "j", false, "Output JSON (for scripting/UI)$")
	rootCmd.AddCommand(statusCmd)
}
