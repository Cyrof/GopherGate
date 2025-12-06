package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
)

type WireGuardService struct {
	gatewayv1.UnimplementedWireGuardServiceServer
}

func NewWireGuardService() *WireGuardService {
	return &WireGuardService{}
}

// CreatePeer handler for gRPC
func (s *WireGuardService) CreatePeer(
	ctx context.Context,
	req *gatewayv1.CreatePeerRequest,
) (*gatewayv1.CreatePeerResponse, error) {

	// map proto -> wgsvc request
	svcReq := wgsvc.CreatePeerRequest{
		Iface:             req.GetIface(),
		Name:              req.GetName(),
		PublicKey:         req.GetPublicKey(),
		AllowedCIDRs:      req.GetAllowedCidrs(),
		Endpoint:          req.GetEndpoint(),
		KeepaliveSeconds:  int(req.GetKeepaliveSeconds()),
		ReplaceAllowedIPs: req.GetReplaceAllowedIps(),
	}

	svcResp, err := wgsvc.CreatePeer(ctx, svcReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create peer: %v", err)
	}

	return &gatewayv1.CreatePeerResponse{
		Iface:         svcReq.Iface,
		Name:          svcReq.Name,
		PublicKey:     svcReq.PublicKey,
		ConfigApplied: svcResp.ConfigApplied,
	}, nil
}

// Delete peer
func (s *WireGuardService) DeletePeer(
	ctx context.Context,
	req *gatewayv1.DeletePeerRequest,
) (*gatewayv1.DeletePeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}
	if req.GetPublicKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "public_key is required")
	}

	// map proto
	svcReq := wgsvc.DeletePeerRequest{
		Iface:     req.GetIface(),
		PublicKey: req.GetPublicKey(),
	}

	svcResp, err := wgsvc.DeletePeer(ctx, svcReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete peer: %v", err)
	}

	return &gatewayv1.DeletePeerResponse{
		Iface:     svcReq.Iface,
		PublicKey: svcReq.PublicKey,
		Removed:   svcResp.Removed,
	}, nil
}
