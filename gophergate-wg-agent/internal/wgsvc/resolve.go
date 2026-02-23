package wgsvc

import (
	"context"
	"fmt"

	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
)

func ResolvePublicKey(ctx context.Context, repo *data.Repository, name, pubkey string) (string, error) {
	if pubkey != "" {
		return pubkey, nil
	}
	if name == "" {
		return "", fmt.Errorf("either --name or --pubkey must be provided")
	}
	key, err := repo.GetPublicKeyByName(ctx, name)
	if err != nil {
		return "", fmt.Errorf("lookup by name %q failed: %w", name, err)
	}
	return key, nil
}
