# GopherGate WireGuard Agent (development progress)

The **Gophergate WireGuard Agent** is a lightweight management service that provides peer and interface CRUD opertaions for WireGuard networks via CLI and gRPC (upcoming).

It forms the backend foundation for the [GopherGate project](https://github.com/Cyrof/GopherGate), handling secure peer provisioning, key registration, and server-side configuration management.

This marks the **first complete functional release** of the agent. At this stage, all core features for peer creation are implemented and stable. Client configuration export (download/QR) will be added in an upcoming release.

## Features

- Peer creation and registration on a live WireGuard interface
- Integrated logging, environment, and path management via `gophergate-core`
- Configuration validation with structured error handling
- Designed for future gRPC integration
- Tested with manual client setup (see below)

## Prerequisites

- Go 1.21+ (matching `go.mod`)
- A configured WireGuard interface (e.g., `wg0`)
- **Root privileges** for network administation (`sudo`) or a binary with the `CAP_NET_ADMIN` capability

## Development Environment

Before running locally, ensure the environment is set up correctly.

#### 1. `.env` file

The agent uses the same environment configuration as defineed in [gophergate-core](https://github.com/Cyrof/GopherGate/blob/dev/gophergate-core/README.md). A `.env` file is needed at the root of the project (or in your working directory) with entries such as:

```ini
GOPHERGATE_ENV=dev
DATABASE_URL=postgres://gg_admin:<your_password>@127.0.0.1:5432/gophergate?sslmode=disable
```

> Refer to the [GopherGate Core README](https://github.com/Cyrof/GopherGate/blob/dev/gophergate-core/README.md) for the latest environment variable reference.

#### 2. Database (development)

A Postgres environment for local testing is provided under [dev-sim](https://github.com/Cyrof/GopherGate/blob/dev/dev-sim/README.md). Follow the instruction in that folder's **README** to spin up the database container before running the agent.

## Running (development)

Ensure you WireGuard server interface exists and is up (e.g., `wg0`). See [wireguard-dev](https://github.com/Cyrof/GopherGate/tree/dev/wireguard-dev) for example setups.

Run the agent in development mode:

```bash
sudo go run ./cmd/gophergate-wg-agent
```

Alternatively, build and grant the capability once (Linux):

```bash
go build -o gophergate-wg-agent ./cmd/gophergate-wg-agent/
sudo setcap cap_net_admin+ep ./gophergate-wg-agent
./gophergate-wg-agent
```

> If you see `logger sync error: sync /dev/stdout: invalid argument` on exit, it's a harmless flush issue with stdout and will be handled in a future update.

## Command Overview

You can view all available commands and flags using:

```bash
sudo go run ./cmd/gophergate-wg-agent --help
```

Example output (abridged):

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
  status      Show WireGuard dsice and peer status

Flags:
  -h, --help   help for gophergate-wg-agent

Use "gophergate-wg-agent [command] --help" for more information about a command.
```

## Manual client setup (temporary, for developers)

Because the agent doesn't export a file yet, developers must manually create a client tunnel and use its **public key** with `create`.

1. Create a new, **empty** tunnel in your WireGuard client and copy its **PublicKey**.
2. Run the `create` command (above) using that public key and an address (e.g., `10.13.13.3/32`).
3. Manually fill the client tunnel configuration as below:

    ```ini
    # Client (developer) tunnel
    [Interface]
    PrivateKey = <CLIENT_PRIVATE_KEY>
    Address = 10.13.13.3/32
    DNS = 10.13.13.1


    [Peer]
    PublicKey = <SERVER_PUBLIC_KEY>
    AllowedIPs = 10.13.13.0/24
    Endpoint = <SERVER_ADDR>:51820
    PersistentKeepalive = 25
    ```

- `<SERVER_ADDR>` is the IP/hostname of the machine running `gophergate-wg-agent`.
- `<SERVER_PUBLIC_KEY` is the server's WireGuard public key (e.g., from `wg show` or your server config).
- Ensure server-side forwarding/NAT is configured if you need egress via the server.

## Notes for contributors

- The project currently runs in **development mode only**; production hardening will follow.
- Uses `gophergate-core` for:
    - Logging (`logx`)
    - Path management (`paths`)
    - Environment loading (`envx`)
    - Datebase and configuration foundations (`dbx`, future use)
- gRPC layer and client configuration generation are under active devlopment.
