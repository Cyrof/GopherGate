# Gophergate-UI

The **GopherGate UI** is the frontend interface for the GopherGate system. It provides a simple, lightweight web interface for managing WireGuard peers and interacting with the GopherGate backend services via gRPC.

## Features

- **Peer Management**: Full CRUD operations for WireGuard peers
  - Create new peers with public keys, allowed IPs, endpoints, and keepalive settings
  - View all peers with connection status
  - Edit peer configurations
  - Delete peers
- **Real-time Data**: Direct integration with `gophergate-wg-agent` via gRPC
- **Health Check**: API endpoint for monitoring service status
- **Structured Logging**: Comprehensive logging using Zap logger
- **Built on gophergate-core**: Standardized paths, logging, and environment configuration

## Prerequisites

- Go 1.21+ (or the version specified in `go.mod`)
- gRPC backend (`gophergate-wg-agent`) running and accessible
- WireGuard interface configured on the system (default: `wg0`)

## Development Setup

#### 1. Create a `.env` file

Inside the **gophergate-ui** directory:

```dotenv
GOPHERGATE_ENV=dev
HTTP_ADDR=:3000
GRPC_ADDR=127.0.0.1:7443
GRPC_TLS_ENABLE=false
WG_IFACE=wg0
```

**Meaning**:

- `GOPHERGATE_ENV`: environment flag shared across GopherGate subsystems.
- `HTTP_ADDR`: address/port the UI will listen on.
- `GRPC_ADDR`: gRPC server address for the agent/backend.
- `GRPC_TLS_ENABLE`: Enable TLS for gRPC connection (default: `false`)
- `WG_IFACE`: WireGuard interface name (default: `wg0`)
    > For more details on environments variables, see the **gophergate-core README**.

### 2. Install Dependencies

```bash
go mod tidy
```

This will download all required dependencies including gRPC packages.

### 3. Start the Backend Agent

Ensure `gophergate-wg-agent` is running and listening on the configured gRPC address. Refer to the agent's [README.md](https://github.com/Cyrof/GopherGate/blob/dev/gophergate-wg-agent/README.md) for the latest start up command.

### 4. Start the UI

From the `gophergate-ui` directory, run:

```bash
go run ./cmd/ui/main.go
```

You should see output similar to:
```
INFO  ui starting  version=0.2.1 http=:3000 grpc=127.0.0.1:7443 env=dev
INFO  grpc.dial    addr=127.0.0.1:7443 tls=false
```

### 5. Access the Interface

Open your browser and navigate to:

- **http://localhost:3000** → Index page
- **http://localhost:3000/peers** → Peer management (CRUD operations)
- **http://localhost:3000/api/health** → Health check endpoint

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Home page |
| GET | `/peers` | List all peers |
| POST | `/peers` | Create new peer |
| GET | `/peers/edit?pubkey=<key>` | Edit form for specific peer |
| POST | `/peers/edit` | Update peer configuration |
| POST | `/peers/delete` | Delete peer |
| GET | `/api/health` | Health check (returns JSON) |


## Notes for Contributors

- This release focuses on **UI structure**, not backend connectivity.
- All CRUD screens are intentionally pre-built so the next PR can plug in gRPC quickly.
- Avoid duplicating configuration details &mdash; `.env`, logging, and paths are standardised in gophergate-core.
- When backend integration is introduced, this README will expand to include the full API flow.
