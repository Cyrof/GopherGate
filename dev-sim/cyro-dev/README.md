# Cyro Dev Services

Shared non-production development services for CyroStack/Coder workspaces.

This folder contains optional Kubernetes manifests for local development from a Coder workspace or any pod inside the cluster.

## Purpose

The `cyro-dev` namespace is used for temporary or non-production services that support development and testing, such as:

- PostgreSQL test database
- Optional WireGuard test instance
- Future dev-only services such as Redis, MinIO, or mock APIs

These services are intended for development only and should not be treated as production workloads.

## Folder Structure

```text
dev-sim/cyro-dev/
├── namespace.yaml
├── postgres.yaml
├── wireguard.yaml
└── README.md
```

## Deploy Namespace

Create the shared dev namespace first:

```bash
kubectl apply -f namespace.yaml
```

Check:

```bash
kubectl get ns cyro-dev
```

## Deploy PostgreSQL

PostgreSQL is the default dev database service.

```bash
kubectl apply -f postgres.yaml
```

Check resources:

```bash
kubectl -n cyro-dev get pods,svc,pvc,secrets
```

Expected service:

```text
cyro-dev-postgres.cyro-dev.svc.cluster.local:5432
```

Default database values:

```text
Database: gophergate
Username: gg_admin
Password: devpassword
```

Backend connection string:

```bash
export DATABASE_URL="postgres://gg_admin:devpassword@cyro-dev-postgres.cyro-dev.svc.cluster.local:5432/gophergate?sslmode=disable"
```

## Deploy WireGuard

WireGuard is optional and should only be deployed when testing VPN-related behaviour.

It requires privileged permissions and host networking, so it should not be left running unless needed.

```bash
kubectl apply -f wireguard.yaml
```

Check resources:

```bash
kubectl -n cyro-dev get pods,svc,pvc
```

View logs:

```bash
kubectl -n cyro-dev logs deploy/cyro-dev-wireguard -f
```

## Coder Development Flow

For full GopherGate development in Coder, deploy all required dev services first:

```bash
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f wireguard.yaml
```

Check that both PostgreSQL and WireGuard are running:

```bash
kubectl -n cyro-dev get pods,svc,pvc
```

Expected services:

```text
cyro-dev-postgres.cyro-dev.svc.cluster.local:5432
cyro-dev-wireguard.cyro-dev.svc.cluster.local:7443
```

The PostgreSQL service is used by both the agent and UI. The WireGuard pod provides the `wg0` interface for VPN-related testing.

## Build Agent Binary from Coder

The agent image is only created during release, so for development we manually build the latest local agent binary and copy it into the WireGuard pod.

From the Coder workspace, navigate to the agent repo:

```bash
cd ~/workspace/GopherGate/gophergate-wg-agent
```

Build the agent binary for the cluster node architecture:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
go build -o /tmp/gophergate-wg-agent ./cmd/gophergate-wg-agent
```

Get the WireGuard pod name:

```bash
POD="$(kubectl -n cyro-dev get pod -l app=cyro-dev-wireguard -o jsonpath='{.items[0].metadata.name}')"
echo "$POD"
```

Copy the built agent binary into the WireGuard pod:

```bash
kubectl -n cyro-dev cp /tmp/gophergate-wg-agent "$POD":/tmp/gophergate-wg-agent -c wireguard
```

Exec into the WireGuard pod:

```bash
kubectl -n cyro-dev exec -it "$POD" -c wireguard -- bash
```

Inside the WireGuard pod, verify that `wg0` exists:

```bash
wg show
ip link show wg0
```

Start the agent gRPC server inside the WireGuard pod:

```bash
chmod +x /tmp/gophergate-wg-agent

DATABASE_URL="postgres://gg_admin:devpassword@cyro-dev-postgres.cyro-dev.svc.cluster.local:5432/gophergate?sslmode=disable" \
GOPHERGATE_ENV=dev \
/tmp/gophergate-wg-agent serve
```

Keep this terminal running while testing the UI.

## Configure UI to Use WireGuard Pod Agent

In another Coder terminal, navigate to the UI repo:

```bash
cd ~/workspace/GopherGate/gophergate-ui
```

Update the UI `.env` file to point to the agent gRPC server running inside the WireGuard pod:

```env
GOPHERGATE_ENV=dev
HTTP_ADDR=:3000
GRPC_ADDR=cyro-dev-wireguard.cyro-dev.svc.cluster.local:7443
GRPC_TLS_ENABLE=false
WG_IFACE=wg0
DATABASE_URL=postgres://gg_admin:devpassword@cyro-dev-postgres.cyro-dev.svc.cluster.local:5432/gophergate?sslmode=disable
```

Then start the UI:

```bash
go run ./cmd/ui/
```

Access the UI through the Coder workspace port for `3000`.

## Development Loop After Agent Code Changes

When agent code changes, rebuild and copy the binary again:

```bash
cd ~/workspace/GopherGate/gophergate-wg-agent

GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
go build -o /tmp/gophergate-wg-agent ./cmd/gophergate-wg-agent

POD="$(kubectl -n cyro-dev get pod -l app=cyro-dev-wireguard -o jsonpath='{.items[0].metadata.name}')"

kubectl -n cyro-dev cp /tmp/gophergate-wg-agent "$POD":/tmp/gophergate-wg-agent -c wireguard
```

Then restart the agent process inside the WireGuard pod.

## Restart PostgreSQL

```bash
kubectl -n cyro-dev rollout restart deploy/cyro-dev-postgres
```

View logs:

```bash
kubectl -n cyro-dev logs deploy/cyro-dev-postgres -f
```

Connect to PostgreSQL:

```bash
kubectl -n cyro-dev exec -it deploy/cyro-dev-postgres -- psql -U gg_admin -d gophergate
```

## Remove PostgreSQL

This removes the Deployment, Service, Secret, and PVC if they are all in `postgres.yaml`.

```bash
kubectl delete -f postgres.yaml
```

Warning: deleting the PVC will remove the dev database data.

## Remove WireGuard

```bash
kubectl delete -f wireguard.yaml
```

This removes WireGuard while keeping PostgreSQL running.

## Remove All Dev Services

```bash
kubectl delete -f wireguard.yaml --ignore-not-found
kubectl delete -f postgres.yaml --ignore-not-found
kubectl delete -f namespace.yaml --ignore-not-found
```

Warning: removing the namespace deletes all resources inside `cyro-dev`.

## Notes

- PostgreSQL is safe to keep running for normal dev use.
- WireGuard is privileged and should be deployed only when needed.
- The agent must run inside the WireGuard pod to access the `wg0` interface.
- Running the agent locally in the Coder terminal will not access the WireGuard pod's `wg0`.
- The agent image is not required for development because the binary is manually built and copied into the WireGuard pod.
- This namespace is non-production.
- Do not store production secrets here.
- Default credentials are for local/dev testing only.
