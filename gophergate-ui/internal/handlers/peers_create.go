package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Create(c *gin.Context) {
	name := normalisePeerName(c.PostForm("name"))
	ip := strings.TrimSpace(c.PostForm("ip"))
	keepaliveStr := c.PostForm("keepalive")
	pubkey := strings.TrimSpace(c.PostForm("pubkey"))
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

	autoAssignIP := ip == ""
	allowedCIDRs := make([]string, 0, 1)
	manualCIDR := ""

	if !autoAssignIP {
		manualCIDR, err = normaliseManualPeerCIDR(ip, existingRows)
		if err != nil {
			p.log.Warnw("peer.create.resolve_manual_ip_failed", "value", ip, "err", err)
			statusCode := http.StatusBadRequest
			if errors.Is(err, errPeerIPAlreadyUsed) {
				statusCode = http.StatusConflict
			}
			p.renderPeersCreateError(c, statusCode, peerAllowedIPErrorMessage(err))
			return
		}
		allowedCIDRs = append(allowedCIDRs, manualCIDR)
	}

	req := &gatewayv2.CreatePeerRequest{
		Iface:             p.wgIface,
		Name:              name,
		PublicKey:         pubkey,
		AllowedCidrs:      allowedCIDRs,
		Endpoint:          endpoint,
		KeepaliveSeconds:  keepalive,
		ReplaceAllowedIps: true,
		AutoAssignIp:      autoAssignIP,
	}

	resp, err := p.grpc.CreatePeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.create.grpc_error", "auto_assign_ip", autoAssignIP, "err", err)
		p.renderPeersCreateError(c, peerErrorHTTPStatus(err), peerActionError("create", err))
		return
	}

	p.log.Infow(
		"peer.create.success",
		"name", name,
		"pubkey", pubkey,
		"auto_assign_ip", autoAssignIP,
		"requested_cidr", manualCIDR,
		"assigned_ip", resp.AssignedIp,
		"assigned_cidr", resp.AssignedCidr,
		"config_applied", resp.ConfigApplied,
	)

	c.Redirect(http.StatusSeeOther, "/peers")
}

func (p *Peers) renderPeersCreateError(c *gin.Context, statusCode int, errorMsg string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rows, err := p.listPeerRows(ctx)
	if err != nil {
		p.log.Warnw("peer.create.reload_existing_peers_failed", "err", err)
		rows = []peer{}
	}

	pool := p.loadIPPoolStatus(ctx)
	data := peerPageData("Peers", rows, "", pool)
	data["openCreateModal"] = true
	data["createError"] = errorMsg
	data["createForm"] = map[string]string{
		"name":      c.PostForm("name"),
		"pubkey":    c.PostForm("pubkey"),
		"ip":        c.PostForm("ip"),
		"endpoint":  c.PostForm("endpoint"),
		"keepalive": c.PostForm("keepalive"),
	}

	c.HTML(statusCode, "peers.tmpl", data)
}
