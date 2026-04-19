package data

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) InsertPeerTrafficSnapshot(ctx context.Context, s PeerTrafficSnapshot) error {
	const q = `
		insert into peer_traffic_snapshots (iface, public_key, rx_bytes, tx_bytes, total_bytes, recorded_at)
		values ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.Exec(ctx, q,
		s.Iface,
		s.PublicKey,
		int64(s.RXBytes),
		int64(s.TXBytes),
		int64(s.TotalBytes),
		s.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("insert peer traffic snapshot: %w", err)
	}
	return nil
}

func (r *Repository) ListPeerTrafficPoints(ctx context.Context, iface, publicKey string, since time.Time) ([]PeerTrafficPoint, error) {
	const q = `
		select recorded_at, rx_bytes, tx_bytes, total_bytes
		from peer_traffic_snapshots
		where iface = $1
		  and public_key = $2
		  and recorded_at >= $3
		order by recorded_at asc;
	`

	rows, err := r.db.Query(ctx, q, iface, publicKey, since)
	if err != nil {
		return nil, fmt.Errorf("list peer traffic points: %w", err)
	}
	defer rows.Close()

	var out []PeerTrafficPoint
	for rows.Next() {
		var p PeerTrafficPoint
		var rx, tx, total int64
		if err := rows.Scan(&p.Timestamp, &rx, &tx, &total); err != nil {
			return nil, fmt.Errorf("scan peer traffic point: %w", err)
		}
		p.RXBytes = uint64(rx)
		p.TXBytes = uint64(tx)
		p.TotalBytes = uint64(total)
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate peer traffic points: %w", err)
	}

	return out, nil
}

func (r *Repository) DeletePeerTrafficOlderThan(ctx context.Context, cutoff time.Time) error {
	const q = `
		delete from peer_traffic_snapshots
		where recorded_at < $1;
	`

	_, err := r.db.Exec(ctx, q, cutoff)
	if err != nil {
		return fmt.Errorf("delete old peer traffic snapshots: %w", err)
	}
	return nil
}
