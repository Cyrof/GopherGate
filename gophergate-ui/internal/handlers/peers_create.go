package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Create(c *gin.Context) {
	name := c.PostForm("name")
	ip := c.PostForm("ip")
	keepaliveStr := c.PostForm("keepalive")
	pubkey := c.PostForm("pubkey")
	endpoint := c.PostForm("endpoint")

	keepalive := int32(0)
	if keepaliveStr != "" {
		val, err := strconv.Atoi(keepaliveStr)
		if err != nil {
			p.log.Warnw("peer.create.invalid_keepalive", "value", keepaliveStr, "err", err)
			c.HTML(http.StatusBadRequest, "peers.tmpl", peerPageData("Peers", []peer{}, "Invalid keepalive value"))
			return
		}
		keepalive = int32(val)
	}

	allowedCIRDs := []string{}
	if ip != "" {
		allowedCIRDs = append(allowedCIRDs, ip)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &gatewayv1.CreatePeerRequest{
		Iface:             p.wgIface,
		Name:              name,
		PublicKey:         pubkey,
		AllowedCidrs:      allowedCIRDs,
		Endpoint:          endpoint,
		KeepaliveSeconds:  keepalive,
		ReplaceAllowedIps: false,
	}

	resp, err := p.grpc.CreatePeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.create.grpc_error", "err", err)
		c.HTML(http.StatusInternalServerError, "peers.tmpl", peerPageData("Peers", []peer{}, "Failed to create peer: "+err.Error()))
		return
	}

	p.log.Infow("peer.create.success", "name", name, "pubkey", pubkey, "config_applied", resp.ConfigApplied)

	c.Redirect(http.StatusSeeOther, "/peers")
}
