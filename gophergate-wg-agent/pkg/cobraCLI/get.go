package cobraCLI

import (
	"context"
	"encoding/json"
	"errors"
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
			return errors.New("database not initialised")
		}

		repo := data.NewRepository(DB)

		pubKey, err := wgsvc.ResolvePublicKey(ctx, repo, getName, getPub)
		if err != nil {
			return err
		}

		peer, err := wgsvc.GetPeerByPublicKey(getIface, pubKey)
		if err != nil {
			return err
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", " ")
			return enc.Encode(peer)
		}

		cmd.Println("Peer:")
		cmd.Printf(" PublicKey: %s\n", peer.PublicKey)
		cmd.Printf(" Endpoint: %s\n", peer.Endpoint)
		cmd.Printf(" AllowedIP: %v\n", peer.AllowedIPs)
		cmd.Printf(" Handshake: %s\n", peer.Handshake)
		cmd.Printf(" RxBytes: %d\n", peer.RxBytes)
		cmd.Printf(" TxBytes: %d\n", peer.TxBytes)
		cmd.Printf(" Keepalive: %s\n", peer.Keepalive)
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
