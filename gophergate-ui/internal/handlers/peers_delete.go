package handlers

import (
	"context"
	"net/http"
	"time"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Delete(c *gin.Context) {
	pubkeyParam := c.PostForm("pubkey")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &gatewayv1.DeletePeerRequest{
		Iface: p.wgIface,
		PublicKey: pubkeyParam,
	}

	resp, err := p.grpc.DeletePeer(ctx, req)
	if err != nil{
		p.log.Infow("peer.delete.grpc_error", "pubkey", pubkeyParam, "err", err)
		c.Redirect(http.StatusSeeOther, "/peers")
	}
	
	if resp.Removed {
		p.log.Infow("peer.delete.success", "pubkey", pubkeyParam)
	} else {
		p.log.Warnw("peer.delete.not_found", "pubkey", pubkeyParam)
	}

	c.Redirect(http.StatusSeeOther, "/peers")
}
