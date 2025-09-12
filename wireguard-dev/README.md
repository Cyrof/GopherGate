# WireGuard Dev Environment (Docker Compose)

This folder contains a **Docker Compose setup** for running a WireGuard container using [linuxserver/wireguard](https://docs.linuxserver.io/images/docker-wireguard) image.
It is intended for **development and simulation only**. The container acts as a backend WireGuard service so that [`gophergate-wg-agent`](../gophergate-wg-agent) can interact with it without needing WireGuard installed on the host.

## Prerequisites
- Docker v28.4.0 and up

Check that Docker is working:
```bash
docker --version
docker compose version
```

## Usage
### 1. Start the container
From inside the `wireguard-dev/` folder:
```bash
docker compose up -d
```
This will:
- Pull the `lscr.io/linuxserver/wireguard:latest` image if not cached.
- Create the container `wireguard-dev`.
- Generate a server config (`wg0.conf`) and a default peer (`devpeer`).
- Mount all generated configs into `./wireguard/config`.

### 2. View logs
```bash
docker logs wireguard-dev
```

You should see startup information and confirmation that the WireGuard server is active.

### 3. Access generated configs
Peer configs are stored under:

```text
./wireguard/config/peer_devpeer/peer_devpeer.conf
```

The server config (`wg0.conf`) is under `./wireguard/config/wg_confs/`.

### 4. Stop the container
To stop and remove the container + network:
```bash
docker compose -f docker-compose.yml down
```

## Configuration Explained
The compose file includes the following environment variables:
| Variable          | Description                                                                                  | Example           |
| ----------------- | -------------------------------------------------------------------------------------------- | ----------------- |
| `PUID` / `PGID`   | User and group IDs inside the container. Ensures files in `./config` are owned by your user. | `1000`            |
| `TZ`              | Timezone for logs.                                                                           | `Asia/Singapore`  |
| `SERVERURL`       | External address clients use to reach the server. For dev, set to `127.0.0.1`.               | `127.0.0.1`       |
| `SERVERPORT`      | UDP port WireGuard listens on. Mapped to host in the `ports:` section.                       | `51820`           |
| `PEERS`           | Comma-separated list of peers to auto-generate.                                              | `devpeer`         |
| `PEERDNS`         | DNS server pushed to peers. With `auto`, it uses the container’s DNS (`10.13.13.1`).         | `auto`            |

**Volumes:**
- `./wireguard/config:/config` → stores keys, configs, and peer files persistently

**Ports:**
- `51820:51820/udp` → forwards host UDP port 51820 to the container (so you can connect locally with WireGuard client).

**Capabilities:**
- `NET_ADMIN` → required so the container can configure network interfaces.
- `SYS_MODULE` → *optional* but it allows the container to load kernel modules if needed.

**Sysctls:**
- `net.ipv4.conf.all.src_valid_mark=1` is required for proper packet marking with WireGuard.

## Notes
- This setup is **local-only**. Do not expose it publicly.
- Use it to test communication between `gophergate-wg-agent` and the WireGuard backend.
- If you want to test a real VPN tunnel, import the generated `peer_devpeer.conf` into a WireGuard client (Windows, iOS, Android) and connect to `127.0.0.1:51820`.

## References
- [LinuxServer.io WireGuard Image](https://docs.linuxserver.io/images/docker-wireguard)
- [WireGuard Documentation](https://www.wireguard.com)