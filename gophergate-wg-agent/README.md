# GopherGate WireGuard Agent (development progress)

WireGuard agent for **GopherGate**. This service will expose peer and interface management (CRUD) over gRPC/CLI. **Current Status**: `create` works against a running WireGuard interface on the server (adds the peer to the kernel). It does **not** yet return a downloadable client config/QR. Local, developer-only workflow for now.

---

## What works today (dev-only)

- CLI command: `create` (server-side; updated the running `wg` interface)
- Logging & paths via `gophergate-core`
- Local-only testing using `sudo` (needs `CAP_NET_ADMIN`)

What is **not** done yet:

- Client export (no downloadable config/QR yet)
- gRPC service
- Read/Update/Delete peers
- Persistence helpers (e.g., writing back to `wg0.conf`)

## Prerequisites

- Go 1.21+ (matching `go.mod`)
- Root privileges for net admin operations during dev (`sudo`), _or_ a setcap'd binary

---

## Running (development)

Ensure you WireGuard server interface exists and is up (e.g., `wg0`). See [wireguard-dev](https://github.com/Cyrof/GopherGate/tree/dev/wireguard-dev).

Run the agent with `sudo` so the process has `CAP_NET_ADMIN`

```bash
sudo go run ./cmd//gophergate-core
```

Alternatively, build and grant the capability once (Linux):

```bash
go build -o gophergate-wg-agent ./cmd/gophergate-wg-agent/
sudo setcap cap_net_admin+ep ./gophergate-wg-agent
./gophergate-wg-agent
```

> If you see `logger sync error: sync /dev/stdout: invalid argument` on exit, it's a harmless flush issue with stdout and will be handled in a future update.

---

## `create` command (current behavior)

The `create` command is intended to register a new WireGuard peer and (eventually) update the server interface. For now it validates inputs and prints according to flags.

Show help:

```bash
sudo go run ./cmd/gophergate-wg-agent/ create --help
```

**Example (dev)** &mdash; add a desktop peer placeholder to `wg0`:

```bash
sudo ./cmd/gophergate-wg-agent create \
    -i wg0 \
    -p <DESKTOP_PUBLIC_KEY> \
    -a 10.13.13.3/32 \
    -k 25
```

**What you get after runnning `create`**:

- Peer with the given public key is added to `wg0` on the server.
- `AllowedIPs` and (if set) `PersistentKeepalive` are applied.

**What you don't get yet**:

- No client download (no `.conf`/QR export).
- No automatic persistence to `wg0.conf`.

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

- Project runs in **dev mode** locally; no production support yet.
- Logging uses `gophergate-core/logx`; prefer structured fields (e.g., `Infow`).
- Paths use `gophergate-core/paths` to resolve logs/config/data directories.
