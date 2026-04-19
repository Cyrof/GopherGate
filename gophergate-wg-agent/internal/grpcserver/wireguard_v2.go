package grpcserver

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/data"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/service"
	"github.com/Cyrof/GopherGate/gophergate-wg-agent/internal/wgsvc"
)

type WireGuardServiceV2 struct {
	gatewayv2.UnimplementedWireGuardServiceServer
	repo         *data.Repository
	dashboardSvc *service.DashboardService
	trafficSvc   *service.TrafficService
}

func NewWireGuardServiceV2(repo *data.Repository) *WireGuardServiceV2 {
	return &WireGuardServiceV2{
		repo:         repo,
		dashboardSvc: service.NewDashboardService(repo),
		trafficSvc:   service.NewTrafficService(repo),
	}
}

func mapPeerStatusToProtoV2(p *wgsvc.PeerStatus, name string) *gatewayv2.PeerStatus {
	if p == nil {
		return nil
	}

	return &gatewayv2.PeerStatus{
		PublicKey:  p.PublicKey,
		Name:       name,
		Endpoint:   p.Endpoint,
		AllowedIps: p.AllowedIPs,
		Handshake:  p.Handshake,
		RxBytes:    p.RxBytes,
		TxBytes:    p.TxBytes,
		Keepalive:  p.Keepalive,
	}
}

func mapDashboardToProto(in *service.DashboardResult) *gatewayv2.GetDashboardResponse {
	if in == nil {
		return &gatewayv2.GetDashboardResponse{}
	}

	out := &gatewayv2.GetDashboardResponse{
		Summary: &gatewayv2.DashboardSummary{
			PeersOnline:          int32(in.Summary.PeersOnline),
			PeersTotal:           int32(in.Summary.PeersTotal),
			PeersChangedLastHour: int32(in.Summary.PeersChangedLastHour),
			TotalRxBytes:         in.Summary.TotalRXBytes,
			TotalTxBytes:         in.Summary.TotalTXBytes,
			TotalTrafficBytes:    in.Summary.TotalTrafficBytes,
			LastSync:             in.Summary.LastSync,
			AgentStatus:          in.Summary.AgentStatus,
			DegradedNodes:        int32(in.Summary.DegradedNodes),
		},
		Notes:  make([]*gatewayv2.DashboardNote, 0, len(in.Notes)),
		Alerts: make([]*gatewayv2.DashboardAlert, 0, len(in.Alerts)),
		Peers:  make([]*gatewayv2.PeerOverview, 0, len(in.Peers)),
	}

	for _, n := range in.Notes {
		out.Notes = append(out.Notes, &gatewayv2.DashboardNote{
			Title: n.Title,
			Body:  n.Body,
			Level: n.Level,
		})
	}

	for _, a := range in.Alerts {
		out.Alerts = append(out.Alerts, &gatewayv2.DashboardAlert{
			Id:       a.ID,
			Title:    a.Title,
			Body:     a.Body,
			Severity: a.Severity,
		})
	}

	for _, p := range in.Peers {
		out.Peers = append(out.Peers, &gatewayv2.PeerOverview{
			PublicKey:     p.PublicKey,
			Name:          p.Name,
			Endpoint:      p.Endpoint,
			AllowedIps:    p.AllowedIPs,
			RxBytes:       p.RXBytes,
			TxBytes:       p.TXBytes,
			LastHandshake: p.LastHandshake,
			Latency:       p.Latency,
			State:         p.State,
			StateReason:   p.StateReason,
		})
	}

	return out
}

func mapPeerTrafficToProto(in *service.PeerTrafficHistoryResult) *gatewayv2.GetPeerTrafficResponse {
	if in == nil {
		return &gatewayv2.GetPeerTrafficResponse{}
	}

	out := &gatewayv2.GetPeerTrafficResponse{
		Traffic: &gatewayv2.PeerTrafficHistory{
			PublicKey: in.PublicKey,
			Name:      in.Name,
			Points:    make([]*gatewayv2.PeerTrafficPoint, 0, len(in.Points)),
		},
	}

	for _, p := range in.Points {
		out.Traffic.Points = append(out.Traffic.Points, &gatewayv2.PeerTrafficPoint{
			Timestamp:  p.Timestamp.UTC().Format(time.RFC3339),
			RxBytes:    p.RXBytes,
			TxBytes:    p.TXBytes,
			TotalBytes: p.TotalBytes,
		})
	}

	return out
}

func (s *WireGuardServiceV2) GetDashboard(
	ctx context.Context,
	req *gatewayv2.GetDashboardRequest,
) (*gatewayv2.GetDashboardResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}

	result, err := s.dashboardSvc.Build(ctx, req.GetIface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build dashboard: %v", err)
	}

	return mapDashboardToProto(result), nil
}

func (s *WireGuardServiceV2) GetPeerTraffic(
	ctx context.Context,
	req *gatewayv2.GetPeerTrafficRequest,
) (*gatewayv2.GetPeerTrafficResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}
	if req.GetPublicKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "public_key is required")
	}

	rangeName := req.GetRange()
	if rangeName == "" {
		rangeName = "24h"
	}

	result, err := s.trafficSvc.GetPeerTrafficHistory(ctx, req.GetIface(), req.GetPublicKey(), rangeName)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unsupported range") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get peer traffic: %v", err)
	}

	return mapPeerTrafficToProto(result), nil
}

func (s *WireGuardServiceV2) CreatePeer(
	ctx context.Context,
	req *gatewayv2.CreatePeerRequest,
) (*gatewayv2.CreatePeerResponse, error) {
	svcReq := wgsvc.CreatePeerRequest{
		Iface:             req.GetIface(),
		Name:              req.GetName(),
		PublicKey:         req.GetPublicKey(),
		AllowedCIDRs:      req.GetAllowedCidrs(),
		Endpoint:          req.GetEndpoint(),
		KeepaliveSeconds:  int(req.GetKeepaliveSeconds()),
		ReplaceAllowedIPs: req.GetReplaceAllowedIps(),
		Repo:              s.repo,
	}

	svcResp, err := wgsvc.CreatePeer(ctx, svcReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create peer: %v", err)
	}

	return &gatewayv2.CreatePeerResponse{
		Iface:         svcReq.Iface,
		Name:          svcReq.Name,
		PublicKey:     svcReq.PublicKey,
		ConfigApplied: svcResp.ConfigApplied,
	}, nil
}

func (s *WireGuardServiceV2) DeletePeer(
	ctx context.Context,
	req *gatewayv2.DeletePeerRequest,
) (*gatewayv2.DeletePeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}
	if req.GetPublicKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "public_key is required")
	}

	svcReq := wgsvc.DeletePeerRequest{
		Iface:     req.GetIface(),
		PublicKey: req.GetPublicKey(),
		Repo:      s.repo,
	}

	svcResp, err := wgsvc.DeletePeer(ctx, svcReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete peer: %v", err)
	}

	return &gatewayv2.DeletePeerResponse{
		Iface:     svcReq.Iface,
		PublicKey: svcReq.PublicKey,
		Removed:   svcResp.Removed,
	}, nil
}

func (s *WireGuardServiceV2) GetPeer(
	ctx context.Context,
	req *gatewayv2.GetPeerRequest,
) (*gatewayv2.GetPeerResponse, error) {
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

	name := ""
	if s.repo != nil {
		if dbPeer, dbErr := s.repo.GetByPublicKey(ctx, req.GetPublicKey()); dbErr == nil && dbPeer != nil {
			name = dbPeer.Name
		}
	}

	return &gatewayv2.GetPeerResponse{
		Peer: mapPeerStatusToProtoV2(peer, name),
	}, nil
}

func (s *WireGuardServiceV2) ListPeer(
	ctx context.Context,
	req *gatewayv2.ListPeerRequest,
) (*gatewayv2.ListPeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}

	peers, err := wgsvc.ListPeers(req.GetIface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list peers: %v", err)
	}

	nameMap := map[string]string{}
	if s.repo != nil {
		publicKeys := make([]string, 0, len(peers))
		for _, p := range peers {
			publicKeys = append(publicKeys, p.PublicKey)
		}

		names, err := s.repo.NamesByPublicKeys(ctx, publicKeys)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "load peer names: %v", err)
		}
		nameMap = names
	}

	out := make([]*gatewayv2.PeerStatus, 0, len(peers))
	for i := range peers {
		out = append(out, mapPeerStatusToProtoV2(&peers[i], nameMap[peers[i].PublicKey]))
	}

	return &gatewayv2.ListPeerResponse{
		Peers: out,
	}, nil
}

func (s *WireGuardServiceV2) UpdatePeer(
	ctx context.Context,
	req *gatewayv2.UpdatePeerRequest,
) (*gatewayv2.UpdatePeerResponse, error) {
	if req.GetIface() == "" {
		return nil, status.Error(codes.InvalidArgument, "iface is required")
	}
	if req.GetPublicKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "public_key is required")
	}

	var keepalivePtr *int
	if ks := req.GetKeepaliveSeconds(); ks != 0 {
		v := int(ks)
		keepalivePtr = &v
	}

	svcReq := wgsvc.UpdatePeerRequest{
		Iface:              req.GetIface(),
		PublicKey:          req.GetPublicKey(),
		SetAllowedCIDRs:    req.GetSetAllowedCidrs(),
		AppendAllowedCIDRs: req.GetAppendAllowedCidrs(),
		Endpoint:           req.GetEndpoint(),
		KeepaliveSeconds:   keepalivePtr,
		Repo:               s.repo,
	}

	svcResp, err := wgsvc.UpdatePeer(ctx, svcReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update peer: %v", err)
	}

	return &gatewayv2.UpdatePeerResponse{
		Iface:     svcResp.Iface,
		PublicKey: svcResp.PublicKey,
		Changed: &gatewayv2.UpdatePeerResponse_Changed{
			AllowedIps: svcResp.Changed.AllowedIPs,
			Endpoint:   svcResp.Changed.Endpoint,
			Keepalive:  svcResp.Changed.Keepalive,
		},
	}, nil
}
