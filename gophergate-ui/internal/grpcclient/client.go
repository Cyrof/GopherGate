package grpcclient

import (
	"context"
	"fmt"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client Wrapper for gRPC connection and service clients
type Client struct {
	conn *grpc.ClientConn
	wg gatewayv1.WireGuardServiceClient
	log *zap.SugaredLogger
}

// Create a new gRPC client connection
func New(addr string, useTLS bool, log *zap.SugaredLogger) (*Client, error){
	var opts []grpc.DialOption

	if useTLS {
		// TODO: Add TLS credentials when needed
		return nil, fmt.Errorf("TLS not yet implemented")
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	log.Infow("grpc.dial", "addr", addr, "tls", useTLS)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("Failed to create gRPC client: %w", err)
	}

	return &Client{
		conn: conn,
		wg: gatewayv1.NewWireGuardServiceClient(conn),
		log: log,
	}, nil
}

// Close the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Create a new WireGuard Peer
func (c *Client) CreatePeer (ctx context.Context, req *gatewayv1.CreatePeerRequest) (*gatewayv1.CreatePeerResponse, error){
	c.log.Infow("grpc.create_peer", "name", req.Name, "pubkey", req.PublicKey)
	return c.wg.CreatePeer(ctx, req)
}

// Retrieves specific peer via public key
func (c *Client) GetPeer(ctx context.Context, req *gatewayv1.GetPeerRequest) (*gatewayv1.GetPeerResponse, error){
	c.log.Infow("grpc.get_peer", "iface", req.Iface, "pubkey", req.PublicKey)
	return c.wg.GetPeer(ctx, req)
}

// Retrieves ALL peers for an interface
func (c *Client) ListPeers(ctx context.Context, req *gatewayv1.ListPeerRequest) (*gatewayv1.ListPeerResponse, error){
	c.log.Infow("grpc.list_peers", "iface", req.Iface)
	return c.wg.ListPeer(ctx, req)
}

// Update an existing peer
func (c *Client) UpdatePeer(ctx context.Context, req *gatewayv1.UpdatePeerRequest) (*gatewayv1.UpdatePeerResponse, error){
	c.log.Infow("grpc.update_peer", "pubkey", req.PublicKey)
	return c.wg.UpdatePeer(ctx, req)
}

// Remove a peer
func (c *Client) DeletePeer(ctx context.Context, req *gatewayv1.DeletePeerRequest) (*gatewayv1.DeletePeerResponse, error){
	c.log.Infow("grpc.delete_peer", "pubkey", req.PublicKey)
	return c.wg.DeletePeer(ctx, req)
}