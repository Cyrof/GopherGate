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

Example backend usage from Coder:

```bash
cd ~/workspace/<repo>
export DATABASE_URL="postgres://gg_admin:devpassword@cyro-dev-postgres.cyro-dev.svc.cluster.local:5432/gophergate?sslmode=disable"
go run .
```

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

## Remove WireGuard

```bash
kubectl delete -f wireguard.yaml
```

This removes WireGuard while keeping PostgreSQL running.

## Typical Coder Dev Flow

1. Start Coder workspace.
2. Clone project repo under `~/workspace`.
3. Deploy dev services:

```bash
kubectl apply -f dev-services/cyro-dev/namespace.yaml
kubectl apply -f dev-services/cyro-dev/postgres.yaml
```

4. Run backend locally inside Coder:

```bash
export DATABASE_URL="postgres://gg_admin:devpassword@cyro-dev-postgres.cyro-dev.svc.cluster.local:5432/gophergate?sslmode=disable"
go run .
```

5. Run frontend locally inside Coder:

```bash
npm install
npm run dev -- --host 0.0.0.0
```

6. Access frontend/backend through Coder workspace ports or Coder apps.

## Notes

- PostgreSQL is safe to keep running for normal dev use.
- WireGuard is privileged and should be deployed only when needed.
- This namespace is non-production.
- Do not store production secrets here.
- Default credentials are for local/dev testing only.