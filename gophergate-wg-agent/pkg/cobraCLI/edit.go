package cobraCLI

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	upIface         string
	upPubKey        string
	upName          string
	upSetAllowed    []string
	upAppendAllowed []string
	upEndpoint      string
	upKeepalive     int
	upKeepaliveSet  bool
	upAsJSON        bool
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"update"},
	Short:   "Update a WireGaurd peer (by public key or name)",
	Long: `Update a WireGuard peer on the running interface. You may:
- Replace or append AllowedIPs
- Set endpoint (host:port)
- Set keepalive seconds (0 disables; omit flag to leave unchanged)

Specify either --pubkey or --name (name resovles via DB). If both are provided, --pubkey takes precedence,`,
	Example: `
	# Replace AllowedIPs
	gophergate-wg-agent edit --pubkey <base64> --set-allowed 10.0.0.2/32 --set-allowed 10.0.1.0/32

	# Append AllowedIPs
	gophergate-wg-agent edit --name alice --append-allowed 10.2.0.0/32

	# Set endpoint and keepalive (30s)
	gophergate-wg-agent edit --name alice --endpoint vpn.example.com:51820 --keepalive 30
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if upPubKey == "" && upName == "" {
			return fmt.Errorf("either --pubkey or --name is required")
		}
		if len(upSetAllowed) > 0 && len(upAppendAllowed) > 0 {
			return fmt.Errorf("provide either --set-allowed or --append-allowed, not both")
		}

		upKeepaliveSet = cmd.Flags().Changed("keepalive")
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		if DB == nil {
			errf(cmd, "Error: database not initialised (required for --name lookup)\n")
			Log.Errorw("edit failed: DB not initialied for name lookup", "iface", upIface, "name", upName)
			return errors.New("database not initialised")
		}

		repo := data.NewRepository(DB)

		pubKey, err := wgsvc.ResolvePublicKey(ctx, repo, upName, upPubKey)
		if err != nil {
			errf(cmd, "Error: failed to resolve name %q: %v\n", upName, err)
			Log.Errorw("resolve name failed", "iface", upIface, "name", upName, "err", err)
			return err
		}

		var kaPtr *int
		if upKeepaliveSet {
			ka := upKeepalive
			kaPtr = &ka
		}

		endpoint := ""
		if cmd.Flags().Changed("endpoint") && strings.TrimSpace(upEndpoint) != "" {
			endpoint = upEndpoint
		}

		req := wgsvc.UpdatePeerRequest{
			Iface:              upIface,
			PublicKey:          pubKey,
			SetAllowedCIDRs:    upSetAllowed,
			AppendAllowedCIDRs: upAppendAllowed,
			Endpoint:           endpoint,
			KeepaliveSeconds:   kaPtr,
			Repo:               repo,
		}

		resp, err := wgsvc.UpdatePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				errf(cmd, "Error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)\n")
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setup cap_net_admin=ep)")
			} else {
				errf(cmd, "ErrorL update peer failed on %s: %v\n", upIface, err)
				Log.Errorw("update peer failed", "iface", upIface, "err", err)
			}
			return err
		}

		if upAsJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			Log.Infow("peer updated", "iface", resp.Iface, "pubKey", resp.PublicKey, "format", "json",
				"change.allowedIPs", resp.Changed.AllowedIPs,
				"change.endpoint", resp.Changed.Endpoint,
				"change.keepalive", resp.Changed.Keepalive,
			)
			return enc.Encode(resp)
		}

		outf(cmd, "Peer updated on %s\n PublicKey: %s\n Changed: allowedIPs=%t endpoint=%t keepalive=%t\n",
			resp.Iface, resp.PublicKey, resp.Changed.AllowedIPs, resp.Changed.Endpoint, resp.Changed.Keepalive)

		Log.Infow("peer updated", "iface", resp.Iface, "pubKey", resp.PublicKey, "format", "text",
			"change.allowedIPs", resp.Changed.AllowedIPs,
			"change.endpoint", resp.Changed.Endpoint,
			"change.keepalive", resp.Changed.Keepalive,
		)

		Log.Debugw("peer update request detail",
			"iface", upIface,
			"pubKey", pubKey,
			"setAllowed", upSetAllowed,
			"appendAllowed", upAppendAllowed,
			"endpoint", upEndpoint,
			"keepaliveSet", upKeepaliveSet,
			"keepalive", upKeepalive,
		)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&upIface, "iface", "i", "wg0", "Wireguard interface name")
	editCmd.Flags().StringVarP(&upPubKey, "pubkey", "p", "", "Peer public key (base64, required)")
	editCmd.Flags().StringVarP(&upName, "name", "n", "", "Peer name (look up in DB)")

	editCmd.Flags().StringSliceVar(&upSetAllowed, "set-allowed", nil, "Replace AllowedIPs with these CIDRs (repeatable)")
	editCmd.Flags().StringSliceVar(&upAppendAllowed, "append-allowed", nil, "Append these CIDRs to AllowedIPs (repeatable)")
	editCmd.Flags().StringVarP(&upEndpoint, "endpoint", "e", "", "Set peer endpoint (host:port)")
	editCmd.Flags().IntVarP(&upKeepalive, "keepalive", "k", 0, "Set keepalive second (0 disables). If ommited, unchanged.")
	editCmd.Flags().BoolVarP(&upAsJSON, "json", "j", false, "Output JSON response")

	rootCmd.AddCommand(editCmd)
}
