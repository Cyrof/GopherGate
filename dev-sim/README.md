# WireGuard Dev Environment (Docker Compose)

This folder contains a **Docker Compose setup** for running a WireGuard container using [linuxserver/wireguard](https://docs.linuxserver.io/images/docker-wireguard) image.
It is intended for **development and simulation only**. The container acts as a backend WireGuard service so that [`gophergate-wg-agent`](../gophergate-wg-agent) can interact with it without needing WireGuard installed on the host.

> **What Changed (Dev QoL)**: We now run the container with **host networking** so the `wg0` interface exists in the host network namespace. This lets a host-running agent use `wgctrl-go` directly. The required `sysctl` is set on the **host** (not inside the container).

## Prerequisites

- Docker v28.4.0 and up
- Linux host with `/dev/net/tun` available
- WireGuard kernel module available (`wireguard`)

Check that Docker is working:

```bash
docker --version
docker compose version
ls -l /dev/net/tun
sudo modprobe wireguard || true # ok if built-in; no output is fine
```

## One-time host sysctl (required)

When using host networking, set this on the **host**:

```bash
# set now
sudo sysctl -w net.ipv4.conf.all.src_valid_mark=1
# make persistent
echo 'net.ipv4.conf.all.src_valid_mark=1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

## Usage

### 1. Start the container

From inside the `wireguard-dev/` folder:

```bash
docker compose -f wireguard-env.yaml up -d
```

This will:

- Pull the `lscr.io/linuxserver/wireguard:latest` image if not cached.
- Create the container `wireguard-dev`.
- Generate a server config (`wg0.conf`) and a default peer (`devpeer`).
- Mount all generated configs into `./wireguard/config`.
- Bring up `wg0` **on the host** (via host networking).

### 2. View logs

```bash
docker logs wireguard-dev
```

You should see startup information and confirmation that the WireGuard server is active.

### 3. Verify WireGuard status (via container &mdash; no host install)

```bash
# show device, peers, handshakes, tx/rx, etc.
docker exec -it wireguard-dev wg show

# show the server config currently applied
docker exec -it wireguard-dev wg showconf wg0
```

(Optionally on the host)

```bash
ip -d link show wg0 # wg0 should exist since we use host networking
```

### 4. Access generated configs

Peer configs are stored under:

```text
# Peer config
./wireguard/config/peer_devpeer/peer_devpeer.conf

# Server config
./wireguard/config/wg_confs/wg0.conf
```

### 5. Stop the container

To stop and remove the container + network:

```bash
docker compose -f wireguard-env.yaml down
```

# PostgreSQL Dev DataBase (for Agent + UI)

Along side the WireGuard container, we run a lightweight Postgres (`postgres:16-alpine`) instance for development.

This DB is used by both the **agent** (peers, keys, configs) and the **UI** (accounts, sessions, cache).

## Usage

### 1. Create secrets

We use Docker secrets instead of a plain `.env` file for passwords. Run this once:

```bash
mkdir -p ./dev-sim/secrets
openssl rand -base64 32 > ./dev-sim/secrets/pg_password.txt
```

This generates a strong random password in `./secrets/pg_password.txt`.

> This should already be included in the `.gitignore` but do check to make sure its never commited.

### 2. Start Postgres

Bring up the dev DB:

```bash
docker compose -f dev-sim.yaml up -d
```

This will:

- Launch the `gg-postgres-dev` container.
- Mount data under `./dev/db_data/` (persistent between restarts).
- Load the password from `secrets/pg_password.txt`.
- Expose Postgres locally on `127.0.0.1:5432`.

### 3. Test Postgres

You can verify the DB is running in a few ways

**Check container health**:

```bash
docker ps --filter name=gg-postgres-dev
```

Status should show `healthy`.

**Check logs**:

```bash
docker logs gg-postgres-dev
```

**Exec into container and check version**:

```bash
docker exec -it gg-postgres-dev psql -U gg_admin -d gophergate -c "SELECT version();"
```

**Healthcheck directly**:

```bash
docker exec gg-postgres-dev pg_isready -U gg_admin -d gophergate
```

**Expected output**:

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

When using host networking, `wg0` comes up in the host namespace.

To ensure packets always source correctly (and `ping` works reliably), edit the generated server config:

```ini
# ./wireguard/config/wg_confs/wg0.conf

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

- In **dev**, the agent runs locally (host terminal) and manages `wg0` directly with `wgctrl-go`.
- There is **no TCP "endpoint"** for the agent; it talks to the kernal via netlink.
- For **Prod (k3s)** later, run the agent as a DaemonSet with `hostNetwork: true` (and typically `CAP_NET_ADMIN`) on the node that has WireGuard.

## Troubleshooting

- `wg show` via `docker exec` **fails**: Ensure the container is running: `docker ps`, check logs: `docker logs wireguard-dev`.
- **No** `wg0` **on host**: Confirm `/dev/net/tun` is presen and mounted, `sudo modprobe wireguard`, and that the host sysctl was applied (`sysctl net.ipv4.conf.all.src_valid_mark`).
- **Can't connect a local WG client**: Ensure UDP/51820 is free and your firewall allows it. Import `peer_devpeer.conf` into a client and connect to `127.0.0.1:51820`.

## References

- [LinuxServer.io WireGuard Image](https://docs.linuxserver.io/images/docker-wireguard)
- [WireGuard Documentation](https://www.wireguard.com)
