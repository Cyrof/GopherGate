package grpcserver

import (
	"context"
	"strings"

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

// Get peer helper
func mapPeerStatusToProto(p *wgsvc.PeerStatus) *gatewayv1.PeerStatus {
	if p == nil {
		return nil
	}

	return &gatewayv1.PeerStatus{
		PublicKey:  p.PublicKey,
		Endpoint:   p.Endpoint,
		AllowedIps: p.AllowedIPs,
		Handshake:  p.Handshake,
		RxBytes:    p.RxBytes,
		TxBytes:    p.TxBytes,
		Keepalive:  p.Keepalive,
	}
}

// Get peer
func (s *WireGuardService) GetPeer(
	ctx context.Context,
	req *gatewayv1.GetPeerRequest,
) (*gatewayv1.GetPeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}
	if req.GetPublicKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "public_key is required")
	}

	peer, err := wgsvc.GetPeerByPublicKey(req.GetIface(), req.GetPublicKey())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "peer not found") {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get peer: %v", err)
	}

	return &gatewayv1.GetPeerResponse{
		Peer: mapPeerStatusToProto(peer),
	}, nil
}

// list peer
func (s *WireGuardService) ListPeer(
	ctx context.Context,
	req *gatewayv1.ListPeerRequest,
) (*gatewayv1.ListPeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}

	peers, err := wgsvc.ListPeers(req.GetIface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list peers: %v", err)
	}

	out := make([]*gatewayv1.PeerStatus, 0, len(peers))
	for i := range peers {
		out = append(out, mapPeerStatusToProto(&peers[i]))
	}

	return &gatewayv1.ListPeerResponse{
		Peers: out,
	}, nil
}
