package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Edit(c *gin.Context) {
	pubkeyOriginal := c.PostForm("pubkey_original")
	ip := c.PostForm("ip")
	keepaliveStr := c.PostForm("keepalive")
	endpoint := c.PostForm("endpoint")

	keepalive := int32(0)
	if keepaliveStr != "" {
		val, err := strconv.Atoi(keepaliveStr)
		if err != nil {
			p.log.Warnw("peer.update.invalid_keepalive", "values", keepaliveStr, "err", err)
			c.Redirect(http.StatusSeeOther, "/peers")
			return
		}
		keepalive = int32(val)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &gatewayv1.UpdatePeerRequest{
		Iface: p.wgIface,
		PublicKey: pubkeyOriginal,
		Endpoint: endpoint,
		KeepaliveSeconds: keepalive,
	}

	if ip != "" {
		req.SetAllowedCidrs = []string{ip}
	}

	resp, err := p.grpc.UpdatePeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.update.grpc_error", "pubkey", pubkeyOriginal, "err", err)
		c.Redirect(http.StatusSeeOther, "/peers")
		return
	}

	p.log.Infow("peer.update.success",
		"pubkey", pubkeyOriginal,
		"changed_allowed_ips", resp.Changed.AllowedIps,
		"changed_endpoint", resp.Changed.Endpoint,
		"changed_keepalive", resp.Changed.Keepalive,
	)
	
	c.Redirect(http.StatusSeeOther, "/peers")
}
