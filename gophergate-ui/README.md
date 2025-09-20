# Gophergate-UI
frontend service for **GopherGate**.

Currently provides:
- **Index page**
- **Health check endpoint**

---

## Prerequisites
- Go 1.21+ (or the version specified in `go.mod`)
- gRPC backend (`gophergate-wg-agent`) running and accessible

---

## Getting Started
1. **Create a  `.env` file** in the `gophergate-ui` folder for development:
    ```text
    GOPHERGATE_ENV=dev
    HTTP_ADDR=:3000
    GRPC_ADDR=127.0.0.1:5051
    ```
    - `GOPHERGATE_ENV`: environment flag shared across GopherGate subsystems.
    - `HTTP_ADDR`: address/port the UI will listen on.
    - `GRPC_ADDR`: gRPC server address for the agent/backend.
2. **Run the server**
    ```bash
    go run .
    ```
3. **Access the UI**
    - http://localhost:3000 -> index page
    - http://localhost:3000/healthz -> health check

---

## Development Checklist
- [x] Index page
- [x] Health check endpoint
- [ ] Peer CRUD UI scaffolding (placeholder only; no gRPC calls yet)
- [ ] Integration with gophergate-wg-agent (separate PR)
- [ ] Minimal CSS for readability (keep styling light)
- [ ] Basic routing/navigation (e.g., Home, Peers placeholder)
- [ ] Error handling for startup and HTTP handlers
- [x] Logging using gophergate-core/logx (console in dev)
- [x] Use gophergate-core/paths if applicable (logs, config)
- [ ] Unit tests for handlers and routing
- [ ] Makefile (run / test / lint)
- [ ] Dockerfile (local dev image)
- [ ] README updates for new endpoints and flags



