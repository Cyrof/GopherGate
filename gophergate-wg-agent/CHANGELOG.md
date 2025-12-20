# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

---

## [v0.4.3] - 2025-12-20

### Added
- (fill)

### Changed
- (fill)

### Fixed
- (fill)

### Removed
- (fill)

## [v0.4.2] - 2025-12-19

### Added
- File-base logging support.
ease
### Changed
- No changes in this release.

### Fixed
- No fixes in this release.

### Removed
- No removals in this release.

## [v0.4.1] - 2025-12-19

### Added
- Updated the **Dockerfile** to build the agent binary dynamically based on target architecture, removing previously hard-coded architecture values.

### Changed
- Resolved cross-architecture build issues that caused image incompatibility on non-matching platforms.

### Fixed
- No additions in this release.

### Removed
- No removals in this release.

## [v0.4.0] - 2025-12-14

### Added

- Added a **Dockerfile** to enable containerised builds and deployments of the `gophergate-wg-agent`.

### Changed

- Updated build and runtime configuration to support container-based execution.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.3.0] - 2025-12-08

### Added

- Integrated **gRPC** support into the agent to enable remote management and inter-service communication within the GopherGate ecosystem.
- Added protobuf service definitions and generated gRPC handlers to support future UI and backend integrations.

### Changed

- Updated project structure to include gRPC-related directories and generated code.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.1] - 2025-11-30

### Added

- Expanded **README** documentation with additional details on CLI usage, available commands, and how the agent interacts with the backend service.

### Changed

- Improved documentation structure and clarity to better guide developers through setup, database usage, and initial configuration.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.0] - 2025-10-26

### Added

- Initial implementation of the **gophergate-wg-agent**, establishing the backend service and CLI for the GopherGate system.
- Implemented full **CRUD operation** for WireGuard peers, enabling creation, retrieval, update, and deletion via the command-line interface.
- Integrated **PostgreSQL support**, including a structured SQL schema directory for database initialisation and management.
- Introduced modular repository layer and service logic to manage peer data and synchronisation between database and WireGuard interfaces.

### Changed

- No changes in this release.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.
