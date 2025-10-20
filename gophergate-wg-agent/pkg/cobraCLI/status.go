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

This uses wgctrl-go (generic netlink). No'wg' CLI is required.`,
	Example: `
	# Show status of wg0 (human-readable logs)
	gophergate-wg-agent status --iface wg0
	
	# Same with shorthand flags
	gophergate-wg-agent status -i wg0
	
	# JSON output for piping to tools (e.g., jp)
	gophergate-wg-agent status -i wg0 -j | jq .
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		ds, err := wgsvc.Status(ctx, statusIface)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				errf(cmd, "Error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)\n")
				Log.Errorw("status permission error", "iface", statusIface, "err", err)
			} else {
				errf(cmd, "Error reading device %s: %v\n", statusIface, err)
				Log.Errorw("status failed", "iface", statusIface, "err", err)
			}
			return err
		}

		if statusAsJson {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			Log.Infow("status ok", "iface", ds.Interface, "listenPort", ds.ListenPort, "peers", len(ds.Peers))
			return enc.Encode(ds)
		}

		// human logs
		outf(cmd, "Device: %s\n ListenPort: %d\n FirewallMark: %d\n PublicKey: %s\n Peers:%d\n\n",
			ds.Interface, ds.ListenPort, ds.FirewallMark, ds.PublicKey, len(ds.Peers),
		)
		for _, p := range ds.Peers {
			handshake := p.Handshake
			if strings.TrimSpace(handshake) == "" {
				handshake = "-"
			}
			outf(cmd, "Peer: %s\n Endpoint: %s\n AllowedIPs: %s\n Handshake: %s\n RxBytes: %d\n TxBytes: %d\n Keepalive: %s\n\n",
				p.PublicKey, p.Endpoint, strings.Join(p.AllowedIPs, ", "), handshake, p.RxBytes, p.TxBytes, p.Keepalive,
			)
		}

		Log.Infow("status displayed", "iface", ds.Interface, "peers", len(ds.Peers))
		for _, p := range ds.Peers {
			Log.Debugw("peer status",
				"iface", ds.Interface,
				"pubKey", p.PublicKey,
				"endpoint", p.Endpoint,
				"allowedIPs", p.AllowedIPs,
				"handshake", p.Handshake,
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
