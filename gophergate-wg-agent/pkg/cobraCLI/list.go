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
	listIface  string
	listOutput string
)

type listItem struct {
	Name       string   `json:"name,omitempty"`
	PublicKey  string   `json:"publicKey"`
	Endpoint   string   `json:"endpoint"`
	AllowedIPs []string `json:"allowedIPs,omitempty"`
	Handshake  string   `json:"handshake,omitempty"`
	RxBytes    uint64   `json:"rxBytes"`
	TxBytes    uint64   `json:"txBytes"`
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List WireGuard peers (joined with DB names)",
	Long: `List peers from the running WireGuard interface and join with peer names from the database.
Use --output json for scripting.`,
	Example: `
	gophergate-wg-agent list
	gophergate-wg-agent list --output json
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		peers, err := wgsvc.ListPeers(listIface)
		if err != nil {
			return err
		}

		pubKeys := make([]string, 0, len(peers))
		for _, p := range peers {
			pubKeys = append(pubKeys, p.PublicKey)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		if DB == nil {
			return errors.New("database not initialised")
		}
		repo := data.NewRepository(DB)
		nameMap, err := repo.NamesByPublicKeys(ctx, pubKeys)
		if err != nil {
			return fmt.Errorf("query names: %w", err)
		}

		rows := make([]listItem, 0, len(peers))
		for _, p := range peers {
			rows = append(rows, listItem{
				Name:       nameMap[p.PublicKey],
				PublicKey:  p.PublicKey,
				Endpoint:   p.Endpoint,
				AllowedIPs: p.AllowedIPs,
				Handshake:  p.Handshake,
				RxBytes:    p.RxBytes,
				TxBytes:    p.TxBytes,
			})
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
			fmt.Fprintf(cmd.OutOrStdout(), "%-16s %-28s %-22s %-12s %-12s %-20s\n", "Name", "PublicKey", "Endpoint", "RxBytes", "TxBytes", "Handshake")
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%-16s %-28s %-22s %-12d %-12d %-20s\n",
					trunc(r.Name, 16),
					truncKey(r.PublicKey, 28),
					trunc(r.Endpoint, 22),
					r.RxBytes,
					r.TxBytes,
					trunc(r.Handshake, 20),
				)
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

func trunc(s string, max int) string {
	if max <= 3 || len(s) <= max {
		if len(s) > max {
			return s[:max]
		}
		return s
	}
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "..."
}

func truncKey(pk string, max int) string {
	if len(pk) <= max {
		return pk
	}
	if max <= 3 {
		return pk[:max]
	}
	front := (max - 1) / 2
	back := max - 1 - front
	if back < 1 {
		back = 1
	}
	return pk[:front] + "..." + pk[len(pk)-back:]
}
