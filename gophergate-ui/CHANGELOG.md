# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

---

## [v0.6.0] - 2026-04-03

### Added

- Integrated Tailwind CSS for utility-first styling
- Added DaisyUI component library for consistent UI components
- Introduced base template layout for shared UI structure across pages

### Changed

- Standardised UI foundation using Tailwind and DaisyUI
- Applied global background design (including mouse-responsive lighting effect) across all pages

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.5.0] - 2026-03-22

### Added

- Added GopherGate logo as tab (favicon) icon

### Changed

- Organised static assets into structured directories (e.g., CSS and images)
- Improved project structure for better maintainability of frontend resources

### Fixed

- No fixes in this release

### Removed

- No removals in this release

## [v0.4.0] - 2025-12-14

### Added

- Added a **Dockerfile** to enable containerised builds and deployments of the `gophergate-ui`

### Changed

- Updated runtime configuration to support container-based execution.

### Fixed

- (fill)

### Removed

- (fill)

## [v0.3.1] - 2025-12-12

### Added

- Updated **README** documentation to include latest startup guide and available API endpoint reference.

### Changed

- Improved documentation clarity to reflect the current gRPC-based UI integration and workflows.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.3.0] - 2025-12-12

### Added

- Integrated **gRPC** support into the UI to enable communication with the `gophergate-wg-agent`.
- Added gRPC client logic to support remote backend operations from the UI.

### Changed

- Updated UI templates and handlers to align with gRPC-based workflows instead of direct or local calls.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.1] - 2025-11-30

### Added

- Updated **README** documentation to include additional details about the current UI structure and setup instructions for development.

### Changed

- Improved overall documentation clarity to reflect the latest project state and upcoming integration with the **gophergate-agent**.

### Fixed

- No fixes in this release.

### Removed

- No removals in this release.

## [v0.2.0] - 2025-10-26

### Added

- Initial implementation of the **gophergate-ui** for Phase 1, establishing the foundational structure of the web interface.
- Basic page routing and functional components implemented to support future integration with **gophergate-agent**.
- Core functionality placeholders created for peer management, system overview, and configuration views.
- Currently serves as a functional prototype with minimal styling and limited usability until backend integration is completed

### Changed

- No changes in this release

### Fixed

- No fixes in this release

### Removed

- No removals in this release.
