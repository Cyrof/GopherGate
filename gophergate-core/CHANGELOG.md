# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

## Usage

To use `gophergate-core` in your Go project, install it as a module dependency:

```bash
go get github.com/Cyrof/GopherGate/gophergate-core@v.0.2.0
```

Then import the packages as needed:

```go
import (
    "github.com/Cyrof/GopherGate/gophergate-core/logger"
    "github.com/Cyrof/GopherGate/gophergate-core/paths"
)
```

---

## [v0.3.3] - 2025-10-13

### Added
- Added **unit tests** and **integration tests** for PostgreSQLto validate core database functionality and connection reliability.
- Introduced a **Makefile** to centralise common development and testing commands, streamlining local and CI workflows.

### Changed
- No functional changes in this release.

### Fixed
- No fixes in this release.

### Removed
- No removals in this release.

## [v0.3.2] - 2025-10-12

### Added
- Implemented a **centralized database module** to provide a unified interface for managing PostgreSQL connections across all dependent services.

### Changed
- Update internal database handling to utilise the centralised connection logic for improved consistency and maintainability.

### Fixed
- No fixes in this release.

### Removed
- No removals in this release.

## [v0.3.1] - 2025-10-05

### Added
- Expanded **README.md** documentation to include detailed usage guides for the new **envx** and **dbx** packages.

### Changed
- No functional changes in this release

### Fixed
- No fixes in this release.

### Removed
- No removals in this release.

## [v0.3.0] - 2025-10-04

### Added

- Introduced **envx** package, a centralised environment loader that manages `.env` configuration and detects development mode settings.
- Added **dbx** package, providing a standardised PostgreSQL connection interface for both local development (via Docker Compose) and production (via containersed deployment on k3s).

### Changed

- Updated **logger** package to utilise the new **envx** module for consistent environment handling and mode detection during initialisation.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.1] - 2025-09-16

### Added

- Implemented `Sync` support in the logger, providing a flush function that can be deferred by individual modules to ensure buffered logs are properly written.

### Changed

- Updated logger initialisation to:
  - Ignore known/benign errors returned during `Sync`.
  - Integrate with **godotenv** to automatically load variables from a `.env` file, allowing the initialisation process to correctly detect development mode from environment configuration.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.0] - 2025-09-16

### Added

- Implemented a centralised **logger package** to provide consistent and structure logging across all core modules.
- Introduced a **default paths package** to establish standardised directory definitions for system-wide usage.

### Changed

- No changes in this release.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.
