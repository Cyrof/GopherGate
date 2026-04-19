# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

---

## [v0.6.0] - 2026-04-19

### Added

- Added backend support for peer traffic history in the v2 flow
- Added RX/TX traffic point handling for UI graphing
- Added a traffic snapshotter to periodically persist peer traffic data to the database
- Added environment variable support for `TRAFFIC_SNAPSHOT_INTERVAL`
- Added environment variable support for `TRAFFIC_RETENTION_DAYS`

### Changed

- Extended the wg-agent backend to support traffic history retrieval for the UI
- Set the default traffic snapshot interval to 1 minute
- Set the default traffic retention period to 3 days

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.5.0] - 2026-04-12

### Added

- Added backend support for protobuf v2 dashboard functionality
- Implemented the required dashboard-related handlers and service logic in the wg-agent

### Changed

- Updated the wg-agent backend to align with the new protobuf v2 definitions
- Extended internal service flow to support dashboard data handling

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.4.6] - 2026-04-06

### Added

- No additions in this release

### Changed

- Minor code cleanup to improve linting compliance

### Fixed

- Fixed code issues that caused linting errors during checks

### Removed

- No removals in this release

## [v0.4.5] - 2026-02-23

### Added

- No additions in this release

### Changed

- Updated documentation for Phase 1 release
- Improved clarity and structure of release documentation

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.4.4] - 2026-02-23

### Added

- No additions for this release.

### Changed

- Refactored create, update, and delete logic for improved database persistence handling.
- Improved internal database interaction structure for better consistency and maintainability.

### Fixed

- Fixed persistence bug where create/update/delete operations were not properly reflected in the database.

### Removed

- No removals in this release.

## [v0.4.3] - 2025-12-20

### Added

- No additions for this release.

### Changed

- Updated runtime to use the kubernetes default/privileged user instead of a custom user.

### Fixed

- No fixes for this release.

### Removed

- Removed custom user configuration in favor of Kubernetes-managed privileges.

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
