# GopherGate WireGuard Agent (development progress)

The **Gophergate WireGuard Agent** is a lightweight management service that provides peer and interface CRUD opertaions for WireGuard networks via CLI and gRPC.

It forms the backend foundation for the [GopherGate project](https://github.com/Cyrof/GopherGate), handling secure peer provisioning, key registration, and server-side configuration management.

This marks the completion of Phase 1 &mdash; all core peer lifecycle operations (create, read, update, delete) are implemented, stable, and persisted to PostgreSQL.

Client configuration export (download/QR generation) will be introduced in a future phase.

## Architecture Overview

The agent supports two operating modes:

### CLI Mode

Direct peer management via Cobra commands.

### gRPC Server Mode

Runs a persistent gRPC service (`serve` command) that allows the GopherGate UI to manage peers remotely.

## Features (Phase 1 Complete)

- Peer creation on a live WireGuard interface
- Peer deletion and cleanup
- Peer updates (AllowedIPs, endpoint, keepalive)
- Peer retrieval (by public key or name)
- Peer listing (joined with PostgreSQL metadata)
- Full PostgreSQL persistence
- Structured error handling
- gRPC server (`serve`) for UI integration
- Shared core utilities via `gophergate-core`
    - Logging (`logx`)
    - Environment loading (`envx`)
    - Path management (`paths`)
    - Database layer (`dbx`)

## Prerequisites

- Go 1.21+ (matching `go.mod`)
- A configured WireGuard interface (e.g., `wg0`)
- **Root privileges** for network administation (`sudo`) or a binary with the `CAP_NET_ADMIN` capability

## Development Environment

Instead of manually configuring WireGuard and Postgres, use the provided devleopment simulation environment at [dev-sim](../dev-sim/)

This folder contains:

- `dev-sim.yaml`
- A local PostgreSQL container
- A WireGuard container
- Setup instructionss in its own README
    > Follow the instructions inside `dev-sim/README` to spin up the full development stack.

This environment is designed specifically for:

- CLI testing
- gRPC testing
- UI integration testing

Before running locally, ensure the environment is set up correctly.

### Environment Configuration

The agent uses the same environment configuration as defined in [gophergate-core](https://github.com/Cyrof/GopherGate/blob/dev/gophergate-core/README.md).

Create a `.env` file in your working directory

```ini
GOPHERGATE_ENV=dev
DATABASE_URL=postgres://gg_admin:<your_password>@127.0.0.1:5432/gophergate?sslmode=disable
```

> Refer to the [GopherGate Core README](https://github.com/Cyrof/GopherGate/blob/dev/gophergate-core/README.md) for the latest environment variable reference.

## Running the Agent (Development)

### CLI Testing Mode

To run CLI commands directly:

```bash
sudo go run ./cmd/gophergate-wg-agent create ...
```

or

```bash
sudo go run ./cmd/gophergate-wg-agent/ list
```

You only need `serve` if testing gRPC.

## gRPC Server Mode (UI Integration)

To start the gRPC server:

```bash
sudo go run ./cmd/gophergate-wg-agent serve
```

This will:

- Connect to PostgreSQL
- Bind to the configured gRPC port
- Allow the GopherGate UI to connect

Once running, you may start the UI and connect it to this agent instance.

> Refer to the [gophergate-ui](https://github.com/Cyrof/GopherGate/tree/dev/gophergate-ui) documentation for UI setup instructions.

## Building Locally (Linux)

```bash
go build -o gophergate-wg-agent ./cmd/gophergate-wg-agent/
sudo setcap cap_net_admin+ep ./gophergate-wg-agent
./gophergate-wg-agent
```

Root or `CAP_NET_ADMIN` is required for WireGuard operations.

## Docker Image (Production Use)

A production-ready agent image is available on Docker Hub:

```
https://hub.docker.com/repository/docker/cyrof/gophergate/general
```

The container:

- Automatically runs the `serve` command
- Is designed to be deployed alongside: - PostgreSQL - gophergate-ui - WireGuard interface
    > This image is intended for integration deployment.

Full production deployment documentation is available in the root GopherGate repository README.

This README focuses only on the agent components.

## Command Overview

You can view all available commands and flags using:

```bash
sudo go run ./cmd/gophergate-wg-agent --help
```

Example output:

```sql
The gophergate-wg-agent provides a simple Cobra-based CLI to manage
WireGuard interfaces and peers. It also exposes a gRPC server to allow
external tools, such as the gophergate-ui, to interact with the WireGuard
service for automation and integration.

Usage:
  gophergate-wg-agent [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create      Create a new WireGuard peer and persist it to the database
  delete      Delete a WireGuard peer (by public key or name)
  edit        Update a WireGaurd peer (by public key or name)
  get         Retrieve details of a WireGuard peer (by public key or name)
  help        Help about any command
  list        List WireGuard peers (joined with DB names)
  serve       Start gRPC server
  status      Show WireGuard dsice and peer status

Flags:
  -h, --help   help for gophergate-wg-agent

Use "gophergate-wg-agent [command] --help" for more information about a command.
```

## Manual Client Setup (Temporary &mdash; Developer Only)

Until configuration export is implemented, client tunnels must be manually created.

### Step 1 &ndash; Generate Client Key

Create a new empty WireGuard tunnel on your client device and copy its PublicKey.

### Step 2 &ndash; Register Peer

```bash
sudo gophergate-wg-agent create \
    --name <name_of_peer> \
    --pubkey <CLIENT_PUBLIC_KEY> \
    --address 10.13.13.3/32 \
    --keepalive 25
```

### Step 3 &ndash; Fill Client Config

```ini
# Client (developer) tunnel
[Interface]
PrivateKey = <CLIENT_PRIVATE_KEY>
Address = 10.13.13.3/32

[Peer]
PublicKey = <SERVER_PUBLIC_KEY>
AllowedIPs = 10.13.13.0/24
Endpoint = <SERVER_ADDR>:51820
PersistentKeepalive = 25
```

- `<SERVER_ADDR>` -> host running the agent
- `<SERVER_PUBLIC_KEY>`-> from `wg show`
- Ensure server-side forwarding/NAT is configured if you need egress via the server.

## Production Notes

- Phase 1 focuses on core peer lifecycle + persistence
- No auto key generation yet
- No config export yet
- No QR generation yet
- Production hardening and security controls will follow in Phase 2+

## Contributor Notes

- Phase 1: Core lifecycle + persistence + gRPC server
- Phase 2: QoL improvements, UI polish, automation
- Future:
    - Auto key generation
    - Config file export
    - QR code generation
    - Role-based access
    - Observability integration
