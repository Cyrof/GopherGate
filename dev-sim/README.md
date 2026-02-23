# WireGuard Dev Environment (Docker Compose)

This folder contains the full development simulation stack for GopherGate.

It provides:

- A WireGuard server (LinuxServer image)
- A PostgreSQL database
- A safe local testing environment for:
    - `gophergate-wg-agent` (CLI mode)
    - `gophergate-wg-agent serve` (gRPC mode)
    - `gophergate-ui`

This setup is strictly for development and local simulation.

## Overview

In development:

- WireGuard runs inside a container using host networking
- PostgreSQL runs in a container
- The agent runs on the host
- The agent talks to:
    - WireGuard via `wgctrl-go` (netlink)
    - PostgreSQL via TCP (`127.0.0.1:5432`)

Because we use host networking, the `wg0` interface exists in the host network namespace.

This allows the agent to control Wireguard without needing WireGuard installed directly on the host.

## Prerequisites

- Docker v28.4.0 and up
- Linux host
- `/dev/net/tun` available
- WireGuard kernel module available (`wireguard`)

Verify:

```bash
docker --version
docker compose version
ls -l /dev/net/tun
sudo modprobe wireguard || true
```

If `modprobe` produces no output, that is fine (module may already be built-in).

## One-time host sysctl (required)

When using host networking, set this on the **host**:

```bash
# set now
sudo sysctl -w net.ipv4.conf.all.src_valid_mark=1
# make persistent
echo 'net.ipv4.conf.all.src_valid_mark=1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

## Starting the Dev Simulation

From inside the `dev-sim/` folder:

```bash
docker compose -f dev-sim.yaml up -d
```

This will:

- Start `wireguard-dev`
- Start `gg-postgres-dev`
- Create `wg0` on the host
- Mount configs under `./wireguard/config`
- Expose Postgres at `127.0.0.1:5432`

## WireGuard Dev Container

### View Logs

```bash
docker logs wireguard-dev
```

### Check WireGuard Status

```bash
# show device, peers, handshakes, tx/rx, etc.
docker exec -it wireguard-dev wg show

# show the server config currently applied
docker exec -it wireguard-dev wg showconf wg0
```

### Verify on Host

Because of host networking:

```bash
ip -d link show wg0
```

You should see `wg0`.

## Generated Config Locations

All configs are mounted under:

```code
./dev-sim/wireguard/config/
```

Peer example:

```code
./dev-sim/wireguard/config/peer_devpeer/peer_devpeer.conf
```

Server config:

```code
# Server config
./dev-sim/wireguard/config/wg_confs/wg0.conf
```

## PostgreSQL Dev Database

Postgres runs using:

```code
postgres:16-alpine
```

Used by:

- gophergate-wg-agent
- gophergate-ui

### Create secrets (Fist Time Only)

```bash
mkdir -p ./dev-sim/secrets
openssl rand -base64 32 > ./dev-sim/secrets/pg_password.txt
```

Ensure `secrets/` is in `.gitignore`.

### Verify Database

Check container:

```bash
docker ps --filter name=gg-postgres-dev
```

Check Logs:

```bash
docker logs gg-postgres-dev
```

Run query:

```bash
docker exec -it gg-postgres-dev psql -U gg_admin -d gophergate -c "SELECT version();"
```

Healthcheck:

```bash
docker exec gg-postgres-dev pg_isready -U gg_admin -d gophergate
```

Expected:

```makefile
gophergate:5432 - accepting connections
```

## Configuration Explained

The compose file includes the following environment variables:
| Variable | Description | Example |
| ----------------- | -------------------------------------------------------------------------------------------- | ----------------- |
| `PUID` / `PGID` | User and group IDs inside the container. Ensures files in `./config` are owned by your user. | `1000` |
| `TZ` | Timezone for logs. | `Asia/Singapore` |
| `SERVERURL` | External address clients use to reach the server. For dev, set to `127.0.0.1`. | `127.0.0.1` |
| `SERVERPORT` | UDP port WireGuard listens on. Mapped to host in the `ports:` section. | `51820` |
| `PEERS` | Comma-separated list of peers to auto-generate. | `devpeer` |
| `PEERDNS` | DNS server pushed to peers. With `auto`, it uses the container’s DNS (`10.13.13.1`). | `auto` |

**Volumes:**

- `./wireguard/config:/config` → stores keys, configs, and peer files persistently
- `/dev/net/tun:/dev/net/tun` → allows WireGuard to create/manage the tunnel device.

**Ports:**

- **Host networking**: _No port mapping needed_. WireGuard listens on `UDP/51820` on the host automatically.
- **If you disable host networking (not recommended for this dev flow)**: expose the port explicitly:
    - `51820:51820/udp` → forwards host UDP 51820 to the container.

**Capabilities:**

- `NET_ADMIN` → required so the container can configure network interfaces.
- `SYS_MODULE` → _optional_ but it allows the container to load kernel modules if needed (often not required if the module is already present on the host).

**Sysctls:**

- `net.ipv4.conf.all.src_valid_mark=1` is required for proper packet marking with WireGuard.
- With **host networking**, Docker cannot set this inside the container. Set it on the **host** instead
- **Also note**: when using host networking, `wg0` comes up in the host namespace.
- Make sure your `wg0.conf` includes a proper subnet route (`10.13.13.0/24`) as shown below in [wg0.conf Routing Notes](#wg0conf-routing-notes-important).

## wg0.conf Routing Notes (Important)

Because we use host networking, `wg0` exists in the host namespace.

Edit:

```code
./dev-sim/wireguard/config/wg_confs/wg0.conf
```

Recommended:

```ini
[Interface]
Address = 10.13.13.1/24
ListenPort = 51820
PrivateKey = <redacted>
Table = off # stop wg-quick from auto-adding routes

PostUp = ip -4 route replace 10.13.13.0/24 dev %i proto static src 10.13.13.1; iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth+ -j MASQUERADE

PostDown = ip -4 route del 10.13.13.0/24 dev %i || true; iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth+ -j MASQUERADE
```

> If you don't need NAT/forwarding (pure dev peer-to-peer testing), you can drop the `iptables` lines and keep only the `ip route` PostUp/PostDown.

### Quick Verficiation

After editing and restarting:

```bash
ip addr show wg0 | grep 10.13.13.1/24
ip -4 route show 10.13.13.0/24
# Expect: 10.13.13.0/24 dev wg0 proto static src 10.13.13.1

ip -4 route get 10.13.13.2
# Expect: dev wg0 ... src 10.13.13.1

ping -c 3 10.13.13.2
```

## How this ties to the agent

In development:

- Agent runs on host
- Uses `wgctrl-go`
- Talks to kernel via netlink
- No TCP endpoint for WireGuard itself
- gRPC is only between UI <-> agent

## Production Note

This setup is strictly for development.

In production:

- The agent runs via Docker image
- Automatically executes `serve`
- Must be deployed alongside:
    - PostgreSQL
    - gophergate-ui
    - A WireGuard-envaled host
- Full deployment instructions are documented in the root GopherGate repository

## Troubleshooting

- `wg show` via `docker exec` **fails**: Ensure the container is running: `docker ps`, check logs: `docker logs wireguard-dev`.
- **No** `wg0` **on host**: Confirm `/dev/net/tun` is presen and mounted, `sudo modprobe wireguard`, and that the host sysctl was applied (`sysctl net.ipv4.conf.all.src_valid_mark`).
- **Can't connect a local WG client**: Ensure UDP/51820 is free and your firewall allows it. Import `peer_devpeer.conf` into a client and connect to `127.0.0.1:51820`.

## References

- [LinuxServer.io WireGuard Image](https://docs.linuxserver.io/images/docker-wireguard)
- [WireGuard Documentation](https://www.wireguard.com)
