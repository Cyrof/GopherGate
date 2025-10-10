package dbx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MigrateConfig struct {
	AdvisoryLockKey int64
	Table           string
}

func MigrateFS(ctx context.Context, pool *pgxpool.Pool, f fs.FS, dir string, cfg *MigrateConfig) error {
	files, err := readSQLFilesSorted(f, dir)
	if err != nil {
		return err
	}
	return migrate(ctx, pool, files, cfg)
}

func MigrateDir(ctx context.Context, pool *pgxpool.Pool, osDir string, cfg *MigrateConfig) error {
	sub := osDir
	return MigrateFS(ctx, pool, dirFS(sub), ".", cfg)
}

type sqlFile struct {
	Version  string
	Name     string
	SQL      string
	Checksum string
}

type osDirFs struct{ root string }

func migrate(ctx context.Context, pool *pgxpool.Pool, files []sqlFile, cfg *MigrateConfig) error {
	table := "schema_migrations"
	if cfg != nil && cfg.Table != "" {
		table = cfg.Table
	}

	return WithTx(ctx, pool, func(ctx context.Context, tx pgxTx) error {
		if cfg != nil && cfg.AdvisoryLockKey != 0 {
			if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", cfg.AdvisoryLockKey); err != nil {
				return fmt.Errorf("advisory lock: %w", err)
			}
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version TEXT PRIMARY KEY,
				checksum TEXT NOT NULL,
				applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`, table)); err != nil {
			return fmt.Errorf("ensure table: %w", err)
		}

		applied := map[string]string{}
		rows, err := tx.Query(ctx, "SELECT version, checksum FROM "+table)
		if err != nil {
			return fmt.Errorf("select applied: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var v, c string
			if err := rows.Scan(&v, &c); err != nil {
				return fmt.Errorf("scan applied: %w", err)
			}
			applied[v] = c
		}
		if rows.Err() != nil {
			return rows.Err()
		}

		// verify checksum for already applied
		for _, f := range files {
			if old, ok := applied[f.Version]; ok {
				if old != f.Checksum {
					return fmt.Errorf("migration %s checksum mismatch (db=%s code=%s)", f.Name, old, f.Checksum)
				}
				continue // already applied and matches
			}
			// run migration
			if _, err := tx.Exec(ctx, f.SQL); err != nil {
				return fmt.Errorf("apply %s: %w", f.Name, err)
			}
			// record
			if _, err := tx.Exec(ctx, "INSERT INTO "+table+"(version, checksum, applied_at) VALUES ($1,$2,$3)", f.Version, f.Checksum, time.Now()); err != nil {
				return fmt.Errorf("record %s: %w", f.Name, err)
			}
		}
		return nil
	})
}

func readSQLFilesSorted(fsys fs.FS, dir string) ([]sqlFile, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("readdir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, errors.New("no .sql files found in " + dir)
	}
	sort.Strings(names)

	var out []sqlFile
	for _, name := range names {
		fp := path.Join(dir, name)
		b, err := fs.ReadFile(fsys, fp)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fp, err)
		}
		sum := sha256.Sum256(b)
		out = append(out, sqlFile{
			Version:  strings.TrimSuffix(name, path.Ext(name)),
			Name:     name,
			SQL:      string(b),
			Checksum: hex.EncodeToString(sum[:]),
		})
	}
	return out, nil
}

func dirFS(root string) fs.FS {
	return osDirFs{root: root}
}

func (d osDirFs) Open(name string) (fs.File, error) {
	return openFile(path.Join(d.root, name))
}

func openFile(p string) (fs.File, error) { return osOpen(p) }

var osOpen = func(p string) (fs.File, error) { return os.Open(p) }

var _ pgxTx = (pgx.Tx)(nil)
