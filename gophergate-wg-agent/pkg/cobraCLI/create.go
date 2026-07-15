package cobraCLI

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/ippool"
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
	createAutoIP     bool
)

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "Create a new WireGuard peer and persist it to the database",
	Long: `The create command provisions a new WireGuard peer, applies the configuration to the
WireGuard interface, and persists the peer record into the database. Each peer must
have a unique public key. The peer address can be supplied manually or allocated from the configured IP pool.`,
	Example: `
	# Create a new peer with basic settings 
	gophergate-wg-agent create --name alice --pubkey <base64> --allowed 10.0.0.2/32 --keepalive 25

	# Create a peer with an automatically assigned address
	gophergate-wg-agent create --name bob --pubkey <base64> --auto-ip --keepalive 25

	# Auto-assign the peer address and add another routed network
	gophergate-wg-agent create -i wg1 -n carol -p <base64> --auto-ip -a 10.0.1.0/24 -k 30
  	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if createKeepalive < 0 {
			return errors.New("--keepalive must be >= 0")
		}
		if !createAutoIP && len(createAllowed) == 0 {
			return errors.New("--allowed is required unless --auto-ip is enabled")
		}
		// validate any extra allowed CIDRs
		for _, cidr := range createAllowed {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("invalid --allowed CIDR %q: %w", cidr, err)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// persist to DB
		if DB == nil {
			errf(cmd, "Error: database no initialied; cannot persist peer\n")
			Log.Errorw("create failed: DB not intialised", "iface", createIface, "name", createName, "pubKey", createPubKey)
			return errors.New("database not initialised")
		}

		repo := data.NewRepository(DB)
		addressPool, err := ippool.FromEnv()
		if err != nil {
			return fmt.Errorf("load IP pool configuration: %w", err)
		}

		req := wgsvc.CreatePeerRequest{
			Iface:             createIface,
			Name:              createName,
			PublicKey:         createPubKey,
			AllowedCIDRs:      createAllowed,
			Endpoint:          createEndpoint,
			KeepaliveSeconds:  createKeepalive,
			ReplaceAllowedIPs: createReplaceIPs,
			AutoAssignIP:      createAutoIP,
			IPPool:            addressPool,
			Repo:              repo,
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 8*time.Second)
		defer cancel()

		resp, err := wgsvc.CreatePeer(ctx, req)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "operation not permitted") {
				errf(cmd, "Error: need CAP_NET_ADMIN (sudo or setcap cap_net_admin=ep)\n")
				Log.Errorw("create permission error", "iface", createIface, "name", createName, "pubKey", createPubKey, "err", err)
			} else {
				errf(cmd, "Error: create peer failed on %s: %v\n", createIface, err)
				Log.Errorw("create peer failed", "iface", createIface, "name", createName, "pubKey", createPubKey, "err", err)
			}
			return err
		}

		outf(cmd, "Peer created on %s\n ID: %s\n Name: %s\n PublicKey: %s\n AssignedIP: %s\n AssignedCIDR: %s\n ConfigApplied: %t\n",
			resp.Iface, resp.ID, resp.Name, resp.PublicKey, resp.AssignedIP, resp.AssignedCIDR, resp.ConfigApplied,
		)

		Log.Infow("peerCreated",
			"id", resp.ID,
			"name", resp.Name,
			"iface", resp.Iface,
			"pubKey", resp.PublicKey,
			"configApplied", resp.ConfigApplied,
			"ip", resp.AssignedIP,
			"autoAssigned", createAutoIP,
		)

		Log.Debugw("peer create request detail",
			"iface", createIface,
			"name", createName,
			"pubKey", createPubKey,
			"allowedCIDRs", createAllowed,
			"endpoint", createEndpoint,
			"keepalive", createKeepalive,
			"replaceIPs", createReplaceIPs,
		)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createIface, "iface", "i", "wg0", "WireGuard interface name")
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Friendly name for this peer (optional)")
	createCmd.Flags().StringVarP(&createPubKey, "pubkey", "p", "", "Peer public key (base64, required)")
	createCmd.Flags().StringSliceVarP(&createAllowed, "allowed", "a", nil, "Additional allowed IPs/routes (CIDR). Required unless --auto-ip is enabled")
	createCmd.Flags().IntVarP(&createKeepalive, "keepalive", "k", 0, "Persistent keepalive in seconds (0 = disabled)")
	createCmd.Flags().BoolVarP(&createReplaceIPs, "replace-ips", "r", true, "Replace existing AllowedIPs for this peer")
	createCmd.Flags().BoolVar(&createAutoIP, "auto-ip", false, "Automatically assign the lowest free IP from the configured pool")

	if err := createCmd.MarkFlagRequired("pubkey"); err != nil {
		Log.Errorw("Failed to mark flag as required", "flag", "pubkey", "error", err)
	}

	if err := createCmd.MarkFlagRequired("keepalive"); err != nil {
		Log.Errorw("Failed to mark flag as required", "flag", "keepalive", "error", err)
	}

	rootCmd.AddCommand(createCmd)
}
