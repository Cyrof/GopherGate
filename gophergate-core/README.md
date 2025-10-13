# GopherGate Core

`gophergate-core` is the **shared utilites module** for the [GopherGate project](https://github.com/Cyrof/GopherGate).

It provides common functionality that both `gophergate-wg-agent` and `gophergate-ui` rely on, ensuring consistency and reducing duplication across applications.

---

### Purpose

- Define **standardised paths** for storing logs, configs, and data for GopherGate applications.
- Provide a **structured logger** (Zap + Lumberjack) with consistent defaults for both development and production.
- Centralise **environment loading** (`.env`) and **database connection management**.
- Provide **built-in database migration support** for all subsystems.
- Act as a foundation for future cross-application utitlies (config parsing, environment helper, etc).

---

### Current Features

#### 1. Paths (`paths` package)

- Creates and manges the standard directory structure under the user's `$XDG_STATE_HOME` or `~/.local/state/gophergate`.
- Ensures application specific folders exist (e.g., `gophergate-core`, `gophergate-wg-agent`, `gophergate-ui`).
- Provide helpers for accessing log file locations.

#### 2. Logger (`logx` package)

- Uses [Uber Zap](https://github.com/uber-go/zap) for structure JSON logging.
- Uses [Lumberjack](https://github.com/natefinch/lumberjack) for log file rotation.
- Auto-detects environment:
  - `GOPHERGATE_ENV=dev` -> colorful console logs (debug-friendly).
  - Default -> JSON logs written to rotating file.

#### 3. Environment (`envx` packages)

- Loads `.env` if present (via [godotenv](https://github.com/joho/godotenv)).
- Detects `GOPHERGATE_ENV` to configure behavior (e.g., dev vs prod).
- Keeps environment handling consistent across subsystems.

#### 4. Database (`dbx` packages)

- Centralise Postgres connection pool based on [pgxpool](https://github.com/jackc/pgx).
- Supports `DATABASE_URL` or manual config.
- Provides a **built-in migration helper** for running SQL migrations from embedded file systems.
- Shared across all GopherGate services.

---

### Example &mdash; using all core packages together

```go
package main

import (
	"context"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-core/envx"
	"github.com/Cyrof/GopherGate/gophergate-core/logx"
	"github.com/Cyrof/GopherGate/gophergate-core/paths"
	"github.com/Cyrof/GopherGate/gophergate-core/dbx"

    "github.com/your/module/internal/schema/migration"
)

func main() {
	// Load .env if present (for development convenience)
	envx.LoadDotenvIfPresent()

	// Ensure standard application paths exist
    // For gophergate-wg-agent use paths.AppAgent
    // For gophergate-ui use paths.AppUI
	p := paths.ForApp(paths.AppAgent)
	if err := p.Ensure(); err != nil {
		panic(err)
	}

	// Initialize logger
	log, flush := logx.Init(logx.Default(paths.AppAgent))
	defer flush()

	log.Infow("starting GopherGate Agent", "env", envx.Current())

	// Initialize database
	cfg := dbx.Default(paths.AppAgent)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, closePool, err := dbx.Open(ctx, cfg)
	if err != nil {
		log.Fatalw("failed to connect database", "err", err)
	}
	defer closePool()

    // Run migration from embedded FS
    mcfg := &dbx.MigrationConfig{AdvisoryLockKey: 4242}
    if err := dbx.MigrateFS(ctx, pool, migration.FS, ".", mcfg); err != nil {
        log.Fatalw("migrate failed", "err", err)
    }

	log.Infow("connected to database")
    // continue setup
}
```

> **Note**:
>
> - Use `paths.AppAgent` when building the gophergate-wg-agent.
> - Use `paths.AppUI` when building the gophergate-ui.
>
> This ensures each subsystem gets its own isolated directory structure and log files.

### Folder Structure Standard
Each GopherGate application that requires database migrations should follow this directory layout:
```bash
internal
└── internal/schema
    └── internal/schema/migrations
        └── internal/schema/migrations/001_init.sql
```

#### migration.go
Create a `migration.go` file in the same folder to embed all `.sql` files into the binary: 
```go
package migration

import "embed"

//go.embed *.sql
var FS embed.FS
```
> This allows all SQL migration files to be pacakged into your binary or Docker image automatically, enduring consistent schema updates across deployments.

---

### Example `.env`

For development:

```dotenv
GOPHERGATE_ENV=dev
DATABASE_URL=postgres://gg_admin:<your_password>@127.0.0.1:5432/gophergate?sslmode=disable
```

For production:

- Set `DATABASE_URL` via environment variable or
- Leave it unset and configure `Host`, `User`, `Password`, `DBName` programmatically before calling `dbx.Open()`.

If you store Postgres passwords in Docker secrets, build the `DATABASE_URL` from those secrets in your container's entrypoint and export it before running the app &mdash; this keeps `dbx` minimal and portable.

---

### Installation

Add `gophergate-core` to your project with:

```bash
go get github.com/Cyrof/GopherGate/gophergate-core@latest
```

Import in code:

```bash
import (
    "github.com/Cyrof/GopherGate/gophergate-core/envx"
    "github.com/Cyrof/GopherGate/gophergate-core/paths"
    "github.com/Cyrof/GopherGate/gophergate-core/logx"
    "github.com/Cyrof/GopherGate/gophergate-core/dbx"
)
```

---

### Future Scope

- Config parsing (YAML/JSON/TOML).
- Environment variable helpers.
- Common gRPC interceptors and middleware.
- Shared error types and constants.

---

### Contributing

- Keep `gophergate-core` **small and focused**: only add functionality that is shared across multiple GopherGate apps.
- Add/update unit tests when extending functionality.
- Document new features in this README.
- Submit PRs with clear commit messages.
