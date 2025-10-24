package data

import (
	"context"
	"fmt"
)

func (r *Repository) DeleteByPublicKey(ctx context.Context, publicKey string) (string, error) {
	const q = `
		delete from peers
		where public_key = $1
		returning id::text;
	`
	var id string
	if err := r.db.QueryRow(ctx, q, publicKey).Scan(&id); err != nil {
		return "", fmt.Errorf("delete peer: %w", err)
	}
	return id, nil
}
