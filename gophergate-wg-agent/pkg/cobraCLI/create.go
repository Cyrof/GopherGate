package cobraCLI

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
	"github.com/spf13/cobra"
)

var (
	createIface      string
	createName       string
	createPubKey     string
	createAllowed    []string
	createEndpoint   string
	createKeepalive  int
	createReplaceIPs bool
)

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "Create a new WireGuard peer (placeholder, no backend yet)",
	Long: `The create command provisions a new WireGuard peer, applies the configuration to the
WireGuard interface, and persists the peer record into the database. Each peer must
have a unique name, public key, and one or more allowed IPs (CIDRs).`,
	Example: `
	# Create a new peer with basic settings 
	gophergate-wg-agent create --name alice --pubkey <base64> --allowed 10.0.0.2/32 --keepalive 25

	# Create another peer on interface wg1 with multiple allowed 	
	gophergate-wg-agent create -i wg1 -n bob -p <base64> -a 10.0.0.3/32 -a 10.0.1.0/32 -k 30
  	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if createKeepalive < 0 {
			return errors.New("--keepalive must be >= 0")
		}
		// validate allowed CIDRs
		for _, cidr := range createAllowed {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("invalid --allowed CIDR %q: %w", cidr, err)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		req := wgsvc.CreatePeerRequest{
			Iface:             createIface,
			Name:              createName,
			PublicKey:         createPubKey,
			AllowedCIDRs:      createAllowed,
			Endpoint:          createEndpoint,
			KeepaliveSeconds:  createKeepalive,
			ReplaceAllowedIPs: createReplaceIPs,
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 8*time.Second)
		defer cancel()

		resp, err := wgsvc.CreatePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				Log.Errorw("permission error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)")
			} else {
				Log.Errorw("create peer failed", "iface", createIface, "err", err)
			}
			return err
		}

		// persist to DB
		if DB == nil {
			return errors.New("database not initialised")
		}

		primaryIP := pickPrimaryIP(createAllowed)
		repo := data.NewRepository(DB)
		id, err := repo.Insert(ctx, data.Peer{
			Name:                createName,
			PublicKey:           createPubKey,
			IPAddress:           primaryIP,
			Endpoint:            optionalString(createEndpoint),
			PersistentKeepalive: optionalI16(createKeepalive),
		})
		if err != nil {
			return fmt.Errorf("insert peer: %w", err)
		}

		Log.Infow("peerCreated",
			"id", id,
			"name", resp.Name,
			"iface", resp.Iface,
			"pubKey", resp.PublicKey,
			"configApplied", resp.ConfigApplied,
			"ip", primaryIP,
		)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createIface, "iface", "i", "wg0", "WireGuard interface name")
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Friendly name for this peer (optional)")
	createCmd.Flags().StringVarP(&createPubKey, "pubkey", "p", "", "Peer public key (base64, required)")
	createCmd.Flags().StringSliceVarP(&createAllowed, "allowed", "a", nil, "Allowed IPs (CIDR). Repeatable (required)")
	createCmd.Flags().IntVarP(&createKeepalive, "keepalive", "k", 0, "Persistent keepalive in seconds (0 = disabled)")
	createCmd.Flags().BoolVarP(&createReplaceIPs, "replace-ips", "r", true, "Replace existing AllowedIPs for this peer")

	if err := createCmd.MarkFlagRequired("pubkey"); err != nil {
		Log.Errorw("Failed to mark flag as required", "flag", "pubkey", "error", err)
	}

	if err := createCmd.MarkFlagRequired("allowed"); err != nil {
		Log.Errorw("Failed to mark flag as required", "flag", "allowed", "error", err)
	}

	if err := createCmd.MarkFlagRequired("keepalive"); err != nil {
		Log.Errorw("Failed to mark flag as required", "flag", "keepalive", "error", err)
	}

	rootCmd.AddCommand(createCmd)
}

func pickPrimaryIP(cidrs []string) net.IP {
	for _, c := range cidrs {
		ip, ipNet, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		ones, bits := ipNet.Mask.Size()
		if (ip.To4() != nil && ones == 32 && bits == 32) || (ip.To4() == nil && ones == 128 && bits == 128) {
			return ip
		}
	}
	return nil
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optionalI16(v int) *int16 {
	if v <= 0 {
		return nil
	}
	x := int16(v)
	return &x
}
