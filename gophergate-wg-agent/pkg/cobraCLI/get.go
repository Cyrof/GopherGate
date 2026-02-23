package cobraCLI

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	getIface string
	getName  string
	getPub   string
	asJSON   bool
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g", "retrieve"},
	Short:   "Retrieve details of a WireGuard peer (by public key or name)",
	Long: `Show live status for a WireGuard peer from the running interface.
You can specify either --pubkey or --name (name resolves via the DB). If both
are provided, --pubkey takes precedences.`,
	Example: `
	# Get by public key
	gophergate-wg-agent get --pubkey <base64>

	# Get by name (resolve via DB to a public key)
	gophergate-wg-agent get --name alice

	# JSON output 
	gophergate-wg-agent get --name alice --json
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if getPub == "" && getName == "" {
			return errors.New("either --pubkey or --name is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		if DB == nil {
			errf(cmd, "Error: database not initialised (required for --name lookup)\n")
			Log.Errorw("get failed: DB not initialised for name lookup", "iface", getIface, "name", getName)
			return errors.New("database not initialised")
		}

		repo := data.NewRepository(DB)

		pubKey, err := wgsvc.ResolvePublicKey(ctx, repo, getName, getPub)
		if err != nil {
			errf(cmd, "Error: failed to resolve name %q: %v\n", getName, err)
			Log.Errorw("resolve name failed", "iface", getIface, "name", getName, "err", err)
			return err
		}

		peer, err := wgsvc.GetPeerByPublicKey(getIface, pubKey)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				errf(cmd, "Error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)\n")
				Log.Errorw("get permission error", "iface", getIface, "pubKey", pubKey, "err", err)
			} else {
				errf(cmd, "Error: failed to read peer on %s: %v\n", getIface, err)
				Log.Errorw("get failed", "iface", getIface, "pubKey", pubKey, "err", err)
			}
			return err
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			Log.Infow("peer fetched", "iface", getIface, "pubKey", peer.PublicKey, "format", "json")
			return enc.Encode(peer)
		}

		out(cmd, "Peer:\n")
		outf(cmd, " PublicKey: %s\n", peer.PublicKey)
		outf(cmd, " Endpoint: %s\n", peer.Endpoint)
		outf(cmd, " AllowedIP: %v\n", peer.AllowedIPs)
		outf(cmd, " Handshake: %s\n", peer.Handshake)
		outf(cmd, " RxBytes: %d\n", peer.RxBytes)
		outf(cmd, " TxBytes: %d\n", peer.TxBytes)
		outf(cmd, " Keepalive: %s\n", peer.Keepalive)

		Log.Infow("peer fetched", "iface", getIface, "pubKey", peer.PublicKey, "format", "text")
		Log.Debug("peer fetched detail",
			"iface", getIface,
			"pubKey", peer.PublicKey,
			"endpoint", peer.Endpoint,
			"allowedIPs", peer.AllowedIPs,
			"handshake", peer.Handshake,
			"rxBytes", peer.RxBytes,
			"txBytes", peer.TxBytes,
			"keepalive", peer.Keepalive,
		)
		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&getIface, "iface", "i", "wg0", "WireGuard interface name")
	getCmd.Flags().StringVarP(&getName, "name", "n", "", "Peer name (looked up in DB)")
	getCmd.Flags().StringVarP(&getPub, "pubkey", "p", "", "Peer public key (base64)")
	getCmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output JSON")

	rootCmd.AddCommand(getCmd)
}
