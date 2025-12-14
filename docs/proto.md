# GopherGate gRPC & Proto Documentation

This document explains the purpose, structure and usage of the Protocol Buffer (`.proto`) files used in the GopherGate project. It also provides instructions for regenerating gRPC stubs and guidance on how to use the generated codes in both `gophergate-wg-agent` (backend) and `gophergate-ui` (frontend).

## Overview

GopherGate uses Protocol Buffers (protobuf) and gRPC to define a type-safe contract between the WireGuard management agent (`gophergate-wg-agent`) and the UI client (`gophergate-ui`). 

The `.proto` files acts as the **single source of truth** for:
- Available RPC endpoints
- Request and response message formats
- Shared data models

This approach ensures:
- Strong compile-time guarantees
- No drift between UI and backend APIs
- Easier long-term maintenance compared to hand-written REST schemas


### Current Scope

The current `v1` proto definitions covers:
- Peer CRUD operations
    - CreatePeer
    - GetPeer
    - ListPeers
    - UpdatePeer
    - DeletePeer
- WireGuard device
- Peer Status Structures

Future services (e.g. streaming, authentication, metrics) can be added in later phases.

## Folder Structure

```text
gophergate-core/
├── api
│   └── proto
│       └── gateway
│           └── v1
│               └── peer.proto
├── pkg
│   └── gen
│       └── gateway
│           └── v1
│               ├── peer_grpc.pb.go
│               └── peer.pb.go
└── Makefile
```

### Explanation
- `api/proto/gateway/v1/*proto`
    - Human-written proto definitions
    - Versioned API contract (`v1`)
- `pkg/gen/gateway/v1`
    - Auto-generated Go code
    - Contains
        - Message structs (`*.pb.go`)
        - gRPC client & server interfaces (`*_grpc.pb.go`)
- `Makefile`
    - Contains the `make proto` target for regeneration

## Regenerating gRPC stubs
Prerequisities
You must have the following tools installed:
1. Protobuf compiler
```bash
protoc --version
```

If missing:
- Arch/ Manjaro
```bash
sudo pacman -S protobuf
```
- Ubuntu / Debian:
```bash
sudo apt install protobuf-compiler
```

2. Go plugins for protoc
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
Ensure Go's bin directory is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Regeneration Command
From `gophergate-core`:
```bash
make proto
```

This will generate:
```text
pkg/gen/gateway/v1/*.pb.go
```

#### When to regenerate
- Only regenerate **when `.proto` files change**
- Generated files **must be committed** to the repository
- UI and agent both depend on the comitted generated code


## Using Generated Code
### In `gophergate-wg-agent` (Server)
The agent **implements** the generated server interface
```go
import (
	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
)

type WireGuardService struct {
	gatewayv1.UnimplementedWireGuardServiceServer
}

// CreatePeer handler for gRPC
func (s *WireGuardService) CreatePeer(
	ctx context.Context,
	req *gatewayv1.CreatePeerRequest,
) (*gatewayv1.CreatePeerResponse, error) {

	// call existing wgsvc login here
}
```

Register the service:
```go
grpcServer := grpc.NewServer()
gatewayv1.RegisterWireGuardServiceServer(grpcServer, &Server{})
```

The backend listens for incoming gRPC requests from the UI.

### In `gophergate-ui` (Client)
The UI uses the generated **client stubs** to call backend RPCs.
```go
import (
	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
)

grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return err
	}
	defer grpcClient.Close()

client := gatewayv1.NewWireGuardServiceClient(grpcClient)

resp, err := client.CreatePeet(ctx, &gatewayv1.CreatePeerRequest{
    Iface: "wg0",
    PublicKey: "abc123",
})
```

The call is automatically serialised, sent to the agent, handled by the backend implementaion, and deserialised into a response.

## Contribution Guidelines
### Backwards Compatibility (v1)
- **Do not modify or renumber existing fields**
- **Do not change field semantics**
- Always **add new fields with new field number**
- Old clients must continue to work with new servers

### Breaking Changes
If a breaking change is required:
- Create a new package version:
    ```swift
    api/proto/gateway/v2/
    ```
- Keep v1 intact
- Allow UI and agent to migrate independently

## Future Extensions
This documentation and structure are designed to scale.

Future additions may include:
- Streaming RPCs
- Authentication / authorisation services
- Audit logging or metrics APIs

This document should be updated as new services are added.

## Summary
- `.proto` files defines the API contract
- Generated code is shared by UI and backend
- Agent implements server interfaces
- UI uses client stubs
- Versioning ensures long-term stability

For questions or changes, please discuss before modifying proto definitions.