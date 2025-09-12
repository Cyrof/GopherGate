# GopherGate Core

`gophergate-core` is the **shared utilites module** for the [GopherGate project](https://github.com/Cyrof/GopherGate).

It provides common functionality that both `gophergate-wg-agent` and `gophergate-ui` rely on, ensuring consistency and reducing duplication across applications.

---

### Purpose

- Define **standardised paths** for storing logs, configs, and data for GopherGate applications.
- Provide a **structured logger** (Zap + Lumberjack) with consistent defaults for both development and production.
- Act as a foundation for future cross-application utitlies (config parsing, environment helper, etc).

---

### Current Features

#### 1. Paths (`paths` package)

- Creates and manges the standard directory structure under the user's `$XDG_STATE_HOME` or `~/.local/state/gophergate`.
- Ensures application specific folders exist (e.g., `gophergate-core`, `gophergate-wg-agent`, `gophergate-ui`).
- Provide helpers for accessing log file locations.

```go
import "github.com/Cyrof/GopherGate/gophergate-core/paths"

func main() {
    d := paths.ForApp(paths.AppAgent) // or paths.AppUI
    _ = d.Ensure() // will require error handling for linting

    println("Logs will be stored at:", d.LogFile)
}
```

---

#### 2. Logger (`logx` package)

- Uses [Uber Zap](https://github.com/uber-go/zap) for structure JSON logging.
- Uses [Lumberjack](https://github.com/natefinch/lumberjack) for log file rotation.
- Auto-detects environment:
  - `GOPHERGATE_END=dev` -> colorful console logs (debug-friendly).
  - Default -> JSON logs written to rotating file.

```go
import (
    "github.com/Cyrof/GopherGate/gophergate-core/logx"
    "github.com/Cyrof/GopherGate/gophergate-core/paths"
)

func main() {
    log := logx.Init(logx.Default(paths.AppAgent))
    defer log.Sync()

    log.Infow("agent started", "version", "0.1.0")
}
```

---

#### Installation

Add `gophergate-core` to your project with:

```bash
go get github.com/Cyrof/GopherGate/gophergate-core@latest
```

Import in code:

```bash
import (
    "github.com/Cyrof/GopherGate/gophergate-core/paths"
    "github.com/Cyrof/GopherGate/gophergate-core/logx"
)
```

---

#### Future Scope

- Config parsing (YAML/JSON/TOML).
- Environment variable helpers.
- Common gRPC interceptors and middleware.
- Shared error types and constants.

---

#### Contributing

- Keep `gophergate-core` **small and focused**: only add functionality that is shared across multiple GopherGate apps.
- Add/update unit tests when extending functionality.
- Document new features in this README.
- Submit PRs with clear commit messages.
