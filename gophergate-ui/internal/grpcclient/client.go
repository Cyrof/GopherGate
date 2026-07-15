package grpcclient

import (
	"context"
	"fmt"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC connection and service clients.
type Client struct {
	conn *grpc.ClientConn
	wgV1 gatewayv1.WireGuardServiceClient
	wgV2 gatewayv2.WireGuardServiceClient
	log  *zap.SugaredLogger
}

// New creates a new gRPC client connection.
func New(addr string, useTLS bool, log *zap.SugaredLogger) (*Client, error) {
	var opts []grpc.DialOption

	if useTLS {
		// TODO: Add TLS credentials when needed.
		return nil, fmt.Errorf("TLS not yet implemented")
	}

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	log.Infow("grpc.dial", "addr", addr, "tls", useTLS)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &Client{
		conn: conn,
		wgV1: gatewayv1.NewWireGuardServiceClient(conn),
		wgV2: gatewayv2.NewWireGuardServiceClient(conn),
		log:  log,
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreatePeer creates a new WireGuard peer using v2.
func (c *Client) CreatePeer(ctx context.Context, req *gatewayv2.CreatePeerRequest) (*gatewayv2.CreatePeerResponse, error) {
	c.log.Infow("grpc.create_peer", "version", "v2", "name", req.Name, "pubkey", req.PublicKey)
	return c.wgV2.CreatePeer(ctx, req)
}

// GetPeer retrieves a specific peer via public key using v2.
func (c *Client) GetPeer(ctx context.Context, req *gatewayv2.GetPeerRequest) (*gatewayv2.GetPeerResponse, error) {
	c.log.Infow("grpc.get_peer", "version", "v2", "iface", req.Iface, "pubkey", req.PublicKey)
	return c.wgV2.GetPeer(ctx, req)
}

// ListPeers retrieves all peers for an interface using v2.
func (c *Client) ListPeers(ctx context.Context, req *gatewayv2.ListPeerRequest) (*gatewayv2.ListPeerResponse, error) {
	c.log.Infow("grpc.list_peers", "version", "v2", "iface", req.Iface)
	return c.wgV2.ListPeer(ctx, req)
}

// UpdatePeer updates an existing peer using v2.
func (c *Client) UpdatePeer(ctx context.Context, req *gatewayv2.UpdatePeerRequest) (*gatewayv2.UpdatePeerResponse, error) {
	c.log.Infow("grpc.update_peer", "version", "v2", "pubkey", req.PublicKey)
	return c.wgV2.UpdatePeer(ctx, req)
}

// DeletePeer removes a peer using v2.
func (c *Client) DeletePeer(ctx context.Context, req *gatewayv2.DeletePeerRequest) (*gatewayv2.DeletePeerResponse, error) {
	c.log.Infow("grpc.delete_peer", "version", "v2", "pubkey", req.PublicKey)
	return c.wgV2.DeletePeer(ctx, req)
}

// GetIPPool retrieves the agent-managed peer IP pool status using v2.
func (c *Client) GetIPPool(ctx context.Context, req *gatewayv2.GetIPPoolRequest) (*gatewayv2.GetIPPoolResponse, error) {
	c.log.Infow("grpc.get_ip_pool", "version", "v2", "iface", req.Iface)
	return c.wgV2.GetIPPool(ctx, req)
}

// GetDashboard retrieves dashboard data using v2.
func (c *Client) GetDashboard(ctx context.Context, req *gatewayv2.GetDashboardRequest) (*gatewayv2.GetDashboardResponse, error) {
	c.log.Infow("grpc.get_dashboard", "version", "v2", "iface", req.Iface)
	return c.wgV2.GetDashboard(ctx, req)
}

func (c *Client) GetPeerTraffic(ctx context.Context, req *gatewayv2.GetPeerTrafficRequest) (*gatewayv2.GetPeerTrafficResponse, error) {
	c.log.Infow("grpc.get_peer_traffic", "version", "v2", "iface", req.Iface, "pubkey", req.PublicKey, "range", req.Range)
	return c.wgV2.GetPeerTraffic(ctx, req)
}
