# Gophergate-UI

The **GopherGate UI** is the frontend interface for the GopherGate system. It provides a simple, lightweight web interface for managing WireGuard peers and interacting with the GopherGate backend services.

This is the **first functional UI release**, containing full CRUD page scaffolding and navigation. Backend integration via gRPC will be added in future release.

## Features (Current Release)

- Index page
- Health check endpoint
- Peer CRUD page
    - _UI-ready but not wired to backend logic yet_
- Basic table layout for peer listing
- Built on `gophergate-core` (paths, logging, envx)

Planned upcoming features:

- Backend integration with `gophergate-wg-agent` via gRPC
- Working Update functionality
- Minimal styling for readability
- Improved routing/navigation
- UI tests + error handling

## Prerequisites

- Go 1.21+ (or the version specified in `go.mod`)
- gRPC backend (`gophergate-wg-agent`) running and accessible

## Development Setup

#### 1. Create a `.env` file

Inside the **gophergate-ui** directory:

```dotenv
GOPHERGATE_ENV=dev
HTTP_ADDR=:3000
GRPC_ADDR=127.0.0.1:5051
```

**Meaning**:

- `GOPHERGATE_ENV`: environment flag shared across GopherGate subsystems.
- `HTTP_ADDR`: address/port the UI will listen on.
- `GRPC_ADDR`: gRPC server address for the agent/backend.
    > For more details on environments variables, see the **gophergate-core README**.

#### 2. Start the UI

```bash
go run ./cmd/gophergate-ui
```

#### 3. Access the Interface

- http://localhost:3000 -> index page
- http://localhost:3000/healthz -> health check
- http://localhost:3000/peers -> peers CRUD pages

Because backend integration is not yet active, the CRUD pages currently display static layouts and placeholder fields.

## Notes for Contributors

- This release focuses on **UI structure**, not backend connectivity.
- All CRUD screens are intentionally pre-built so the next PR can plug in gRPC quickly.
- Avoid duplicating configuration details &mdash; `.env`, logging, and paths are standardised in gophergate-core.
- When backend integration is introduced, this README will expand to include the full API flow.
