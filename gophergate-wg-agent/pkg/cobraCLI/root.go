package cobraCLI

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	DB      *pgxpool.Pool
	stop    func()
	rootCmd = &cobra.Command{
		Use:   "gophergate-wg-agent",
		Short: "A ClI agent for managing WireGuard with Cobra and gRPC.",
		Long: `The gophergate-wg-agent provides a simple Cobra-based CLI to manage
WireGuard interfaces and peers. It also exposes a gRPC server to allow
external tools, such as the gophergate-ui, to interact with the WireGuard
service for automation and integration.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if DB != nil {
			return nil
		}

		ctx := withShutdown(context.Background())

		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			errf(cmd, "Warning: DATABASE_URL is not set; DB-backend features may be limited.\n")
			Log.Warn("DATABASE_URL not set; continuing without DB")
			return nil
		}

		pool, cleanup, err := dbx.Open(ctx, dbx.Config{DSN: dsn})
		if err != nil {
			errf(cmd, "Error: failed to open database: %v\n", err)
			Log.Errorw("open db failed", "err", err)
			return fmt.Errorf("open db: %w", err)
		}

		DB = pool
		stop = cleanup
		Log.Infow("database initialised")

		// bootstrap WG from DB
		iface := os.Getenv("WG_IFACE")
		if iface == "" {
			iface = "wg0"
		}
		if err := syncWireGuardFromDB(ctx, DB, iface); err != nil {
			Log.Errorw("boostrap sync from DB failed", "iface", iface, "err", err)
		}

		return nil
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if stop != nil {
			stop()
			stop = nil
			Log.Debug("database connection closed")
		}
	}
}

func syncWireGuardFromDB(ctx context.Context, pool *pgxpool.Pool, iface string) error {
	repo := data.NewRepository(pool)

	sctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	recs, err := repo.BootstrapPeers(sctx)
	if err != nil {
		return fmt.Errorf("bootstrap list: %w", err)
	}
	if len(recs) == 0 {
		Log.Infow("no peers in DB to sync", "iface", iface)
		return nil
	}

	cli, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("wgctrl new: %w", err)
	}
	defer func() { _ = cli.Close() }()

	dev, err := cli.Device(iface)
	if err != nil {
		return fmt.Errorf("read device %q: %w", iface, err)
	}
	current := make(map[string]wgtypes.Peer)
	for _, p := range dev.Peers {
		current[p.PublicKey.String()] = p
	}

	var updates []wgtypes.PeerConfig
	var unchanged, changed int

	for _, r := range recs {
		pub, err := wgtypes.ParseKey(r.PublicKey)
		if err != nil {
			Log.Errorw("invalid public key; skipping", "name", r.Name, "pubKey", r.PublicKey, "err", err)
			continue
		}

		// desired allowed ip
		var wantAllowed []net.IPNet
		for _, cidr := range r.Allowed {
			_, ipn, err := net.ParseCIDR(cidr)
			if err != nil {
				Log.Errorw("invalid CIDR; skipping", "name", r.Name, "pubKey", r.PublicKey, "cidr", cidr)
				continue
			}
			wantAllowed = append(wantAllowed, *ipn)
		}

		// endpoint
		var wantEP *net.UDPAddr
		if r.Endpoint != nil && strings.TrimSpace(*r.Endpoint) != "" {
			addr, err := net.ResolveUDPAddr("udp", *r.Endpoint)
			if err != nil {
				Log.Errorw("invalid endpoint; skipping endpoint", "name", r.Name, "pubKey", r.PublicKey, "endpoint", *r.Endpoint, "err", err)
			} else {
				wantEP = addr
			}
		}

		// keepalive
		var wantKA *time.Duration
		if r.Keepalive != nil && *r.Keepalive > 0 {
			d := time.Duration(*r.Keepalive) * time.Second
			wantKA = &d
		}

		// diff against kernel
		cur, has := current[pub.String()]
		need := wgtypes.PeerConfig{PublicKey: pub}

		if !has {
			Log.Debugw("peer missing in kernel, will add", "name", r.Name, "pubKey", r.PublicKey)
			need.ReplaceAllowedIPs = len(wantAllowed) > 0
			need.AllowedIPs = wantAllowed
			need.Endpoint = wantEP
			need.PersistentKeepaliveInterval = wantKA
			updates = append(updates, need)
			changed++
			continue
		}

		// compoase allowedIPs
		if !allowedEqual(cur.AllowedIPs, wantAllowed) {
			need.ReplaceAllowedIPs = len(wantAllowed) > 0
			need.AllowedIPs = wantAllowed
		}

		// compare endpoint
		if !udpAddrEqual(cur.Endpoint, wantEP) {
			need.Endpoint = wantEP
		}

		// compare keepalive
		if !keepaliveEqual(cur.PersistentKeepaliveInterval, wantKA) {
			need.PersistentKeepaliveInterval = wantKA
		}

		// collect only if something changes
		if need.ReplaceAllowedIPs || need.Endpoint != nil || need.PersistentKeepaliveInterval != nil {
			updates = append(updates, need)
			changed++
		} else {
			unchanged++
		}
	}

	if len(updates) == 0 {
		Log.Infow("bootstrap sync: no changes needed", "iface", iface, "unchanged", unchanged)
		return nil
	}
	if err := cli.ConfigureDevice(iface, wgtypes.Config{Peers: updates}); err != nil {
		return fmt.Errorf("configure device: %w", err)
	}
	Log.Infow("bootstrap sync applied", "iface", iface, "changed", changed, "unchanged", unchanged)
	return nil
}

func allowedEqual(cur []net.IPNet, want []net.IPNet) bool {
	if len(cur) != len(want) {
		return false
	}
	m := make(map[string]int, len(cur))
	for _, a := range cur {
		m[cidrString(a)]++
	}
	for _, b := range want {
		k := cidrString(b)
		if m[k] == 0 {
			return false
		}
		m[k]--
	}
	return true
}

func cidrString(n net.IPNet) string {
	return (&net.IPNet{IP: n.IP, Mask: n.Mask}).String()
}

func udpAddrEqual(a, b *net.UDPAddr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.String() == b.String()
}

func keepaliveEqual(cur time.Duration, want *time.Duration) bool {
	if want == nil {
		return cur == 0
	}
	return cur == *want
}

func withShutdown(ctx context.Context) context.Context {
	c, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c.Done()
		cancel()
	}()
	return c
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		Log.Errorw("unexpected error occured", "err", err)
		os.Exit(1)
	}
}
