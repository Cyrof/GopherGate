package handlers

import (
	"context"
	"net/http"
	"time"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/gin-gonic/gin"
)

func (p *Peers) EditForm(c *gin.Context) {
	pubkey := c.Query("pubkey")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &gatewayv2.GetPeerRequest{
		Iface:     p.wgIface,
		PublicKey: pubkey,
	}

	resp, err := p.grpc.GetPeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.edit_form.grpc_error", "pubkey", pubkey, "err", err)
		c.String(http.StatusNotFound, "peer not found")
		return
	}

	ip := ""
	if len(resp.Peer.AllowedIps) > 0 {
		ip = resp.Peer.AllowedIps[0]
	}

	rec := peer{
		Name: resp.Peer.Name,
		PublicKey: resp.Peer.PublicKey,
		IP:        ip,
		Keepalive: resp.Peer.Keepalive,
		Endpoint:  resp.Peer.Endpoint,
		RxBytes:   resp.Peer.RxBytes,
		TxBytes:   resp.Peer.TxBytes,
		Handshake: resp.Peer.Handshake,
	}

	p.renderPeerEditPage(c, http.StatusOK, rec, "")
}

func (p *Peers) renderPeerEditPage(c *gin.Context, status int, rec peer, errorMsg string) {
	c.HTML(status, "peers_edit.tmpl", gin.H{
		"title": "Edit Peer",
		"pageTitle": "Edit Peer",
		"activeNav": "peers",
		"statusText": "Peer Management",
		"error": errorMsg,
		"peer": rec,
	})
}
