package data

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/ippool"
	"github.com/jackc/pgx/v5"
)

type AutoInsertResult struct {
	ID           string
	AssignedIP   netip.Addr
	AssignedCIDR string
	AllowedIPs   []net.IPNet
}

// InsertAutoAssigned serialises allocation for a pool, chooses the lowest free
// address, and inserts the peer in the same transaction.
func (r *Repository) InsertAutoAssigned(ctx context.Context, p Peer, pool *ippool.Pool) (AutoInsertResult, error) {
	if pool == nil {
		return AutoInsertResult{}, fmt.Errorf("IP pool is not configured")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AutoInsertResult{}, fmt.Errorf("begin IP allocation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	lockName := "gophergate-ip-pool:" + pool.CIDR() + ":" + pool.Start().String() + ":" + pool.End().String()
	if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended($1, 0))`, lockName); err != nil {
		return AutoInsertResult{}, fmt.Errorf("lock IP pool: %w", err)
	}

	used, err := ipAddressesInRange(ctx, tx, pool.Start(), pool.End())
	if err != nil {
		return AutoInsertResult{}, err
	}
	assigned, err := pool.NextAvailable(used)
	if err != nil {
		return AutoInsertResult{}, err
	}

	assignedIP := net.IP(assigned.AsSlice())
	assignedNet := net.IPNet{IP: assignedIP, Mask: net.CIDRMask(32, 32)}
	p.IPAddress = assignedIP
	p.AllowedIPs = appendHostCIDR(p.AllowedIPs, assignedNet)

	id, err := insertPeer(ctx, tx, p)
	if err != nil {
		return AutoInsertResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AutoInsertResult{}, fmt.Errorf("commit IP allocation: %w", err)
	}

	return AutoInsertResult{
		ID:           id,
		AssignedIP:   assigned,
		AssignedCIDR: pool.HostCIDR(assigned),
		AllowedIPs:   p.AllowedIPs,
	}, nil
}

func (r *Repository) IPAddressesInRange(ctx context.Context, start, end netip.Addr) (map[netip.Addr]struct{}, error) {
	return ipAddressesInRange(ctx, r.db, start, end)
}

type rowsQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func ipAddressesInRange(ctx context.Context, q rowsQuerier, start, end netip.Addr) (map[netip.Addr]struct{}, error) {
	const stmt = `
		select host(ip_address)
		from peers
		where ip_address is not null
		  and family(ip_address) = 4
		  and ip_address >= $1::inet
		  and ip_address <= $2::inet;
	`
	rows, err := q.Query(ctx, stmt, start.String(), end.String())
	if err != nil {
		return nil, fmt.Errorf("query allocated IPs: %w", err)
	}
	defer rows.Close()

	used := make(map[netip.Addr]struct{})
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan allocated IP: %w", err)
		}
		addr, err := netip.ParseAddr(raw)
		if err != nil || !addr.Is4() {
			continue
		}
		used[addr.Unmap()] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read allocated IPs: %w", err)
	}
	return used, nil
}

func appendHostCIDR(existing []net.IPNet, host net.IPNet) []net.IPNet {
	target := host.String()
	for _, current := range existing {
		if current.String() == target {
			return existing
		}
	}
	return append(existing, host)
}
