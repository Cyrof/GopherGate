package handlers

import (
	"context"
	"net/http"
	"time"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Create(c *gin.Context) {
	name := normalisePeerName(c.PostForm("name"))
	ip := c.PostForm("ip")
	keepaliveStr := c.PostForm("keepalive")
	pubkey := c.PostForm("pubkey")
	endpoint := normaliseEndpoint(c.PostForm("endpoint"))

	keepalive, err := parseKeepaliveSeconds(keepaliveStr)
	if err != nil {
		p.log.Warnw("peer.create.invalid_keepalive", "value", keepaliveStr, "err", err)
		p.renderPeersCreateError(c, http.StatusBadRequest, "Invalid keepalive value. Please enter a number of seconds, for example 25.")
		return
	}

	if err := validateEndpoint(endpoint); err != nil {
		p.log.Warnw("peer.create.invalid_endpoint", "value", endpoint, "err", err)
		p.renderPeersCreateError(c, http.StatusBadRequest, "Invalid endpoint. Leave it blank for roaming peers, or use the format IP:port, for example 192.168.1.100:51820.")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	existingRows, err := p.listPeerRows(ctx)
	if err != nil {
		p.log.Warnw("peer.create.load_existing_peers_failed", "err", err)
		p.renderPeersCreateError(c, http.StatusServiceUnavailable, "Unable to check existing peers before creating a new peer. Please try again.")
		return
	}

	if peerNameExists(existingRows, name) {
		p.renderPeersCreateError(c, http.StatusConflict, "Peer name already exists. Please use a unique name for this peer.")
		return
	}

	allowedCIDR, err := p.normaliseCreateAllowedCIDR(ip, existingRows)
	if err != nil {
		p.log.Warnw("peer.create.resolve_allowed_ip_failed", "value", ip, "err", err)
		p.renderPeersCreateError(c, http.StatusBadRequest, peerAllowedIPErrorMessage(err))
		return
	}

	req := &gatewayv2.CreatePeerRequest{
		Iface:             p.wgIface,
		Name:              name,
		PublicKey:         pubkey,
		AllowedCidrs:      []string{allowedCIDR},
		Endpoint:          endpoint,
		KeepaliveSeconds:  keepalive,
		ReplaceAllowedIps: false,
	}

	resp, err := p.grpc.CreatePeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.create.grpc_error", "err", err)
		p.renderPeersCreateError(c, peerErrorHTTPStatus(err), peerActionError("create", err))
		return
	}

	p.log.Infow(
		"peer.create.success",
		"name", name,
		"pubkey", pubkey,
		"allowed_cidr", allowedCIDR,
		"config_applied", resp.ConfigApplied,
	)

	c.Redirect(http.StatusSeeOther, "/peers")
}

func (p *Peers) renderPeersCreateError(c *gin.Context, status int, errorMsg string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rows, err := p.listPeerRows(ctx)
	if err != nil {
		p.log.Warnw("peer.create.reload_existing_peers_failed", "err", err)
		rows = []peer{}
	}

	data := peerPageData("Peers", rows, errorMsg)
	data["openCreateModal"] = true

	c.HTML(status, "peers.tmpl", data)
}
