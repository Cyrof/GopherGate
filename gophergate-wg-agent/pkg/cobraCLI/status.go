package cobraCLI

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var (
	statusIface  string
	statusAsJson bool
)

type statusPeer struct {
	PublicKey  string   `json:"public_key"`
	Endpoint   string   `json:"endpoint"`
	AllowedIPs []string `json:"allowed_ips"`
	Handshake  string   `json:"latest_handshake"`
	RxBytes    uint64   `json:"rx_bytes"`
	TxBytes    uint64   `json:"tx_bytes"`
	Keepalive  string   `json:"keepalive"`
}

type statusOut struct {
	Interface    string       `json:"interface"`
	ListenPort   int          `json:"listen_port"`
	FirewallMark int          `json:"firewall_mark"`
	PublicKey    string       `json:"public_key"`
	Peers        []statusPeer `json:"peers"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show WireGuard device and peer status",
	Long: `Show current WireGuard device information (listen port, public key, firewall mark)
and all peers (endpoint, allowed IPs, last handshake, bytes, keepalive).

In dev, this reads the device create by the linuxserver/wireguard container (host networking)
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
		cli, err := wgctrl.New()
		if err != nil {
			Log.Errorw("wgctrl init error", "err", err)
			return err
		}
		defer func() {
			if err := cli.Close(); err != nil && Log != nil {
				Log.Warnw("wgctrl close error", "err", err)
			}
		}()

		dev, err := cli.Device(statusIface)
		if err != nil {
			Log.Errorw("failed to read device", "iface", statusIface, "err", err)
			return err
		}

		if statusAsJson {
			out := statusOut{
				Interface:    dev.Name,
				ListenPort:   dev.ListenPort,
				FirewallMark: dev.FirewallMark,
				PublicKey:    dev.PublicKey.String(),
			}
			for _, p := range dev.Peers {
				var ips []string
				for _, ipn := range p.AllowedIPs {
					ips = append(ips, ipn.String())
				}
				ka := ""
				if p.PersistentKeepaliveInterval > 0 {
					ka = p.PersistentKeepaliveInterval.String()
				}
				out.Peers = append(out.Peers, statusPeer{
					PublicKey:  p.PublicKey.String(),
					Endpoint:   fmt.Sprint(p.Endpoint),
					AllowedIPs: ips,
					Handshake:  p.LastHandshakeTime.Format(time.RFC3339),
					RxBytes:    uint64(p.ReceiveBytes),
					TxBytes:    uint64(p.TransmitBytes),
					Keepalive:  ka,
				})
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "	")
			return enc.Encode(out)
		}

		// human logs
		Log.Infow("wg device",
			"name", dev.Name,
			"listenPort", dev.ListenPort,
			"fwMark", dev.FirewallMark,
			"pubKey", dev.PublicKey.String(),
			"peers", len(dev.Peers),
		)
		for _, p := range dev.Peers {
			var ips []string
			for _, ipn := range p.AllowedIPs {
				ips = append(ips, ipn.String())
			}
			ka := ""
			if p.PersistentKeepaliveInterval > 0 {
				ka = p.PersistentKeepaliveInterval.String()
			}
			Log.Infow("peer",
				"pubKey", p.PublicKey.String(),
				"endpoint", p.Endpoint,
				"allowedIPs", strings.Join(ips, ", "),
				"latestHandshake", p.LastHandshakeTime,
				"rxBytes", p.ReceiveBytes,
				"txBytes", p.TransmitBytes,
				"keepalive", ka,
			)
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusIface, "iface", "i", "wg0", "WireGuard interface name")
	statusCmd.Flags().BoolVarP(&statusAsJson, "json", "j", false, "Output JSON (for scripting/UI)")
	rootCmd.AddCommand(statusCmd)
}
