# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

## Usage

To use `gophergate-core` in your Go project, install it as a module dependency:

```bash
go get github.com/Cyrof/GopherGate/gophergate-core
```

Then import the packages as needed:

```go
import (
    "github.com/Cyrof/GopherGate/gophergate-core/logger"
    "github.com/Cyrof/GopherGate/gophergate-core/paths"
    "github.com/Cyrof/GopherGate/gophergate-core/envx"
    "github.com/Cyrof/GopherGate/gophergate-core/dbx"
    gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
)
```

---

## [v0.7.1] - 2026-07-15

### Added
- No additions in this release

### Changed
- Changed `allowed_cidrs` to `cidr` in IPPoolStatus struct

### Fixed
- No fixes in this release

### Removed
- No removals in this release

## [v0.7.0] - 2026-07-13

### Added

- Added v2 IP-pool status messages and the `GetIPPool` RPC
- Added `auto_assign_ip` to peer creation requests
- Added committed address fields to create responses and peer read models

### Changed

- Extended the shared v2 contract for wg-agent automatic address allocation and future frontend integration

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.6.1] - 2026-05-15

### Added

- Added protobuf definitions for peer page functionality
- Added the required, response, and service structures for peer management features

### Changed

- Extended core protobuf v2 definitions to support peer page integration
- Updated shared interface contracts for upcoming peer-related backend and UI functionality

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.6.0] - 2026-04-19

### Added

- Added protobuf definitions for peer traffic point data to support UI graphing
- Added the required message structures for traffic history and peer traffic visualisation

### Changed

- Extended the core protobuf contract to support traffic graph data for the UI
- Updated shared interface definitions to align with dashboard traffic history requirements

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.5.0] - 2026-04-12

### Added

- Added protobuf v2 definitions for dashboard-related backend support
- Introduced the required message and service structure for upcoming dashboard functionality

### Changed

- Expanded the core protobuf contract to support the new dashboard data flow
- Updated shared information definitions to align with backend requirements for v2

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.4.3] - 2025-12-20

### Added

- Kubernetes-specific logging environment override to force stdout logging for containerised deployments.

### Changed

- No changes for this release.

### Fixed

- No fixes for this release.

### Removed

- No removals for this release.

## [v0.4.2] - 2025-12-14

### Added

- No additions in this release.

### Changed

- Updated **README.md** to include minor explaination of gRPC and proto files.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.4.1] - 2025-11-19

### Added

- Updated **README** documentation to include the latest gRPC integration details, directory structure (`api`, `pkg`), and Makefile usage for protobuf regeneration.

### Changed

- Improved general documentation clarity to better reflect the current module architecture and setup instructions.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.4.0] - 2025-11-02

### Added

- Integrated **gRPC** support into the core module to enable efficient inter-service communication within the GopherGate ecosystem.
- Added an **API** directory containing the protobuf (`.proto`) definitions for service interfaces.
- Introduced a **pkg** directory to store auto-generated gRPC and protobuf files, maintaining a clean separation between source definitions and generated artifacts.

### Changed

- Updated **Makefile** to include new targets for protobuf code generation, streamlining regeneration and ensuring consistent build workflows.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.3.5] - 2025-10-13

### Added

- Updated **documentation** to include missing sections required for initialising and running the database component, ensuring clearer setup and configuration guidance for developers.

### Changed

- No functional changes in this release.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.3.4] - 2025-10-13

### Added

- Expanded **documentation** to cover the updated **dbx** module, which now supports automatic database migrations without requiring manual migration setup in individual applications.

### Changed

- No functional changes in this release.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.3.3] - 2025-10-13

### Added

- Added **unit tests** and **integration tests** for PostgreSQL to validate core database functionality and connection reliability.
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
