# Changelog

All notable changes to this project will be documented in this file

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/)

---

## [v0.7.3] - 2026-04-06

### Added

- No additions in this release

### Changed

- Updated the Dockerfile to correct the container build process

### Fixed

- Fixed broken Dockerfile that cause build issues

### Removed

- No removals in this release

## [v0.7.2] - 2026-04-06

### Added

- Added `npm run build:css` to the Docker build process

### Changed

- Updated the Dockerfile to ensure CSS assets are built during image creation

### Fixed

- Fixed missing CSS build step in the Dockerfile

### Removed

- No removals in this release

## [v0.7.1] - 2026-04-05

### Added

- No additionals in this release

### Changed

- Updated runtime stage base image from `alpine:3.20` to `alpine:3.21` for latest security patches

### Fixed

- Fixed CI/CD build failure by updating `gophergate-ui` Dockerfile base image from `golang:1.24-alpine` to `goland:1.25-alpine` to match `go.mod` requirement of `go >= 1.25.0`

### Removed

- No removals in this release

## [v0.7.0] - 2026-04-05

### Added

- Complete session-based authentication system using `gin-contrib/sessions` and `bcrypt`
- Default admin user auto-generation on first startup with random password
- `internal/auth/session.go` for session middleware and helpers
- `internal/auth/seed.go` for admin seeding logic
- `internal/handlers/login.go` with login POST and logout handlers
- `migrations/001_create_users.sql` to create `users` table in PostgreSQL
- `migrations/embed.go` for embedded SQL migrations
- Auth-protected routes using `auth.RequireAuth()` middleware
- Error message display on login page for failed attempts
- `SESSION_SECRET` environment variable for session cookie signing
- Database connection, migration, and admin seeding in application startup

### Changed

- Updated login page form to POST to `/login` endpoint
- Updated `internal/web/server.go` to include session middleware and route protection
- Updated `internal/config/config.go` with database and session configuration
- Updated `cmd/ui/main.go` with database initialization and admin seeding flow
- Dashboard and peer management routes now require authentication
- Unauthenticated users are redirected to the login page

### Fixed

- No fixes in this release

### Removed

- No removals in this release

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
