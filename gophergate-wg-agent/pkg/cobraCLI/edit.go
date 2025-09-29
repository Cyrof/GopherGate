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
	upIface         string
	upPubKey        string
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
		if upPubKey == "" {
			return fmt.Errorf("--pubkey is required")
		}
		if len(upSetAllowed) > 0 && len(upAppendAllowed) > 0 {
			return fmt.Errorf("provide either --set-allowed or --append-allowed, not both")
		}

		upKeepaliveSet = cmd.Flags().Changed("keepalive")
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var kaPtr *int
		if upKeepaliveSet {
			ka := upKeepalive
			kaPtr = &ka
		}

		req := wgsvc.UpdatePeerRequest{
			Iface:              upIface,
			PublicKey:          upPubKey,
			SetAllowedCIDRs:    upSetAllowed,
			AppendAllowedCIDRs: upAppendAllowed,
			Endpoint:           upEndpoint,
			KeepaliveSeconds:   kaPtr,
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		resp, err := wgsvc.UpdatePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setup cap_net_admin=ep)")
			} else {
				Log.Errorw("update peer failed", "iface", upIface, "err", err)
			}
			return err
		}

		if upAsJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(resp)
		}

		Log.Infow("peer updated", "iface", resp.Iface, "pubKey", resp.PublicKey, "change.allowedIPs", resp.Changed.AllowedIPs, "change.endpoint", resp.Changed.Endpoint, "change.keepalive", resp.Changed.Keepalive)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&upIface, "iface", "i", "wg0", "Wireguard interface name")
	editCmd.Flags().StringVarP(&upPubKey, "pubkey", "p", "", "Peer public key (base64, required)")
	editCmd.Flags().StringSliceVar(&upSetAllowed, "set-allowed", nil, "Replace AllowedIPs with these CIDRs (repeatable)")
	editCmd.Flags().StringSliceVar(&upAppendAllowed, "append-allowed", nil, "Append these CIDRs to AllowedIPs (repeatable)")
	editCmd.Flags().StringVarP(&upEndpoint, "endpoint", "e", "", "Set peer endpoint (host:port)")
	editCmd.Flags().IntVarP(&upKeepalive, "keepalive", "k", 0, "Set keepalive second (0 disables). If ommited, unchanged.")
	editCmd.Flags().BoolVarP(&upAsJSON, "json", "j", false, "Output JSON response")

	rootCmd.AddCommand(editCmd)
}
