package gophergatecore_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Cyrof/GopherGate/gophergate-core/dbx"
)

type pgRes struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
	dsn      string
	hostport string
}

func startPostgresWithDockertest(t *testing.T) *pgRes {
	t.Helper()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "connect to docker")

	// Use random host port (HostPort: "")
	runOpts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16", // use debian-based for fewer quirks
		Env: []string{
			"POSTGRES_PASSWORD=pw",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=gg",
		},
	}

	resource, err := pool.RunWithOptions(runOpts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err, "run postgres")

	// Cleanup hook
	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	// Figure out the mapped port
	// var hostPort string
	// {
	// 	port, ok := resource.GetPort("5432/tcp")
	// 	require.True(t, ok)
	// 	hostPort = port // e.g., "32778"
	// }
	pool.MaxWait = 90 * time.Second
	err = pool.Retry(func() error {
		exitCode, execErr := resource.Exec(
			[]string{"pg_isready", "-U", "postgres", "-d", "gg"},
			dockertest.ExecOptions{},
		)
		if execErr != nil {
			return execErr
		}
		if exitCode != 0 {
			return fmt.Errorf("pg_isready exit=%d", exitCode)
		}
		return nil
	})
	require.NoError(t, err, "postgres should be ready (pg_isready)")

	hostPort := resource.GetPort("5432/tcp")
	require.NotEmpty(t, hostPort, "mapped port for 5432/tcp should not be empty")

	addr := net.JoinHostPort("127.0.0.1", hostPort)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "postgres", "pw", addr, "gg")

	{
		conn, dialErr := net.DialTimeout("tcp", addr, 3*time.Second)
		require.NoError(t, dialErr, "tcp connect to Postgres failed")
		_ = conn.Close()
	}

	// Retry until PG is actually ready
	err = pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cfg, e := pgxpool.ParseConfig(dsn)
		if e != nil {
			return e
		}
		p, e := pgxpool.NewWithConfig(ctx, cfg)
		if e != nil {
			return e
		}
		defer p.Close()
		return p.Ping(ctx)
	})
	require.NoError(t, err, "postgres should be ready")

	return &pgRes{pool: pool, resource: resource, dsn: dsn, hostport: hostPort}
}

func openPool(t *testing.T, dsn string) (*pgxpool.Pool, func()) {
	t.Helper()
	cfg := dbx.Default("gg-core-itest")
	cfg.DSN = dsn
	cfg.ConnectTimeout = 5 * time.Second

	ctx := context.Background()
	pool, closeFn, err := dbx.Open(ctx, cfg)
	require.NoError(t, err)
	return pool, closeFn
}

// Reset DB between subtests (fast): drop and recreate the public schema.
func resetDB(t *testing.T, dsn string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer p.Close()
	_, err = p.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err)
}

func TestDBX_DockertestIntegration(t *testing.T) {
	pg := startPostgresWithDockertest(t)

	// Base schema using MigrateFS
	{
		pool, closeFn := openPool(t, pg.dsn)
		defer closeFn()
		ctx := context.Background()
		mfs := fstest.MapFS{
			"001_init.sql": {Data: []byte(`
				create table if not exists it_people(
					id serial primary key,
					name text not null
				);`)},
			"002_seed.sql": {Data: []byte(`
				insert into it_people(name) values ('seed') on conflict do nothing;`)},
		}
		require.NoError(t, dbx.MigrateFS(ctx, pool, mfs, ".", &dbx.MigrateConfig{
			Table: "schema_migrations",
		}))
	}

	t.Run("WithTx commit & rollback", func(t *testing.T) {
		// ensure clean slate for the subtest
		resetDB(t, pg.dsn)
		// re-apply base schema
		pool, closeFn := openPool(t, pg.dsn)
		defer closeFn()
		ctx := context.Background()
		_ = dbx.MigrateFS(ctx, pool, fstest.MapFS{
			"001_init.sql": {Data: []byte(`create table it_people(id serial primary key, name text not null);`)},
			"002_seed.sql": {Data: []byte(`insert into it_people(name) values ('seed') on conflict do nothing;`)},
		}, ".", &dbx.MigrateConfig{Table: "schema_migrations"})

		require.NoError(t, dbx.WithTx(ctx, pool, func(ctx context.Context, tx dbx.Tx) error {
			_, err := tx.Exec(ctx, `insert into it_people(name) values ($1)`, "Ada")
			return err
		}))
		_ = dbx.WithTx(ctx, pool, func(ctx context.Context, tx dbx.Tx) error {
			_, err := tx.Exec(ctx, `insert into it_people(name) values ($1)`, "Grace")
			require.NoError(t, err)
			return fmt.Errorf("force rollback")
		})

		var n int
		require.NoError(t, dbx.WithTx(ctx, pool, func(ctx context.Context, tx dbx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from it_people`).Scan(&n)
		}))
		require.Equal(t, 2, n) // seed + Ada
	})

	t.Run("MigrateDir real files", func(t *testing.T) {
		resetDB(t, pg.dsn)

		pool, closeFn := openPool(t, pg.dsn)
		defer closeFn()
		ctx := context.Background()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "001_a.sql"),
			[]byte(`create table if not exists it_dir_table(id int primary key);`), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "002_b.sql"),
			[]byte(`insert into it_dir_table(id) values (42) on conflict do nothing;`), 0644))

		require.NoError(t, dbx.MigrateDir(ctx, pool, dir, &dbx.MigrateConfig{
			Table:           "schema_migrations_dir",
			AdvisoryLockKey: 22222,
		}))

		var n int
		require.NoError(t, dbx.WithTx(ctx, pool, func(ctx context.Context, tx dbx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from it_dir_table`).Scan(&n)
		}))
		require.Equal(t, 1, n)
	})

	t.Run("Checksum mismatch", func(t *testing.T) {
		resetDB(t, pg.dsn)

		pool, closeFn := openPool(t, pg.dsn)
		defer closeFn()
		ctx := context.Background()

		mfs := fstest.MapFS{
			"010_alpha.sql": {Data: []byte(`create table if not exists it_alpha(x int primary key);`)},
			"020_beta.sql":  {Data: []byte(`insert into it_alpha(x) values (7) on conflict do nothing;`)},
		}
		require.NoError(t, dbx.MigrateFS(ctx, pool, mfs, ".", &dbx.MigrateConfig{
			Table: "schema_migrations_alpha",
		}))

		mfsBad := fstest.MapFS{
			"010_alpha.sql": {Data: []byte(`-- changed
				create table if not exists it_alpha(x int primary key, y int);`)},
			"020_beta.sql": {Data: []byte(`insert into it_alpha(x) values (7) on conflict do nothing;`)},
		}
		err := dbx.MigrateFS(ctx, pool, mfsBad, ".", &dbx.MigrateConfig{
			Table: "schema_migrations_alpha",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "checksum mismatch")
	})
}
