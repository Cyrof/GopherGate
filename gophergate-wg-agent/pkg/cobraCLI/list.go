package cobraCLI

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	listIface  string
	listOutput string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all WireGuard peers (placeholder, no backend yet)",
	Long: `The list commmand will be used to retrieve and display all configured
WireGuard peers, including their names, IDs, and assigned IP addresses.
At present, this command is only a placeholder - the WireGuard service
integration is not yet implemented.`,
	Example: `
	# List all peers
	gophergate-wg-agent list

	# List all peers in JSON format (future flag support)
	gophergate-wg-agent list --output json
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		peers, err := wgsvc.ListPeers(listIface)
		if err != nil {
			return err
		}

		switch strings.ToLower(listOutput) {
		case "json":
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(peers)
		case "table":
			if len(peers) == 0 {
				cmd.Println("No peers found.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-45s %-22s %-12s %-12s\n", "PublicKey", "Endpoint", "RxBytes", "TxBytes")
			for _, p := range peers {
				fmt.Fprintf(cmd.OutOrStdout(), "%-45s %-22s %-12d %-12d\n", p.PublicKey, p.Endpoint, p.RxBytes, p.TxBytes)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s (use table|json)", listOutput)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&listIface, "iface", "i", "wg0", "Wireguard interface name")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "table", "Output format (table|json)")

	rootCmd.AddCommand(listCmd)
}
