package cobraCLI

import (
	"encoding/json"
	"errors"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	getIface string
	asJSON   bool
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g", "retrieve"},
	Short:   "Retrieve details of a WireGuard peer (placeholder, no backend yet)",
	Long: `The get command will be used to fetch details of a WireGuard peer,
such as public key, IP address, and connection status. At present, this
command is only a placeholder - the WireGuard service integration is not
not yet implemented, so running it will not return any peer information`,
	Example: `
	# Retrieve details for a specific peer by name
	gophergate-wg-agent get --name alice

	# Retrieve details for a peer by ID
	gophergate-wg-agent get --id 123
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// make sure either name or id is provided
		if name == "" && peerID == "" {
			return errors.New("you must specify either --name or --id")
		}
		if name != "" && peerID != "" {
			return errors.New("please specify only one of --name or --id, not both")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		peer, err := wgsvc.GetPeerByNameOrID(getIface, name, peerID)
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
	getCmd.Flags().StringVar(&name, "name", "", "Name of the user peer")
	getCmd.Flags().StringVar(&peerID, "id", "", "Unique ID of the peer")
	getCmd.Flags().StringVarP(&getIface, "iface", "i", "wg0", "WireGuard interface name")
	getCmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output JSON")

	rootCmd.AddCommand(getCmd)
}
