package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SeedDefaultAdmin creates a default admin user if non exists.
// The auto-generated password is logged ONCE. The admin should change it.
func SeedDefaultAdmin(ctx context.Context, pool *pgxpool.Pool, log *zap.SugaredLogger) error {
	var count int64
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("check admin exists: %w", err)
	}

	if count > 0 {
		log.Infow("admin user already exists, skipping seed")
		return nil
	}

	password, err := generateRandomPassword(20)
	if err != nil {
		return fmt.Errorf("generate password: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		"admin", string(hashed), "admin",
	)
	if err != nil {
		return fmt.Errorf("insert admin: %w", err)
	}

	log.Infow("==================================")
	log.Infow("  DEFAULT ADMIN CREDENTIALS CREATED")
	log.Infow("  Username: admin")
	log.Infow("  Password: " + password)
	log.Infow("  !! SAVE THIS - it will NOT be shown again")
	log.Infow("==================================")

	return nil
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
