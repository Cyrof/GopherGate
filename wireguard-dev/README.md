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

