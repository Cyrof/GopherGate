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
	endpoint := c.PostForm("endpoint")

	keepalive, err := parseKeepaliveSeconds(keepaliveStr)
	if err != nil {
		p.log.Warnw("peer.create.invalid_keepalive", "value", keepaliveStr, "err", err)
		p.renderPeersCreateError(c, http.StatusBadRequest, "Invalid keepalive value. Please enter a number of seconds, for example 25.")
		return
	}

	if err := validateAllowedCIDR(ip); err != nil {
		p.log.Warnw("peer.create.invalid_allowed_ip", "value", ip, "err", err)
		p.renderPeersCreateError(c, http.StatusBadRequest, "Invalid Allowed IP. Please enter a valid CIDR value, for example 10.13.13.2/32.")
		return
	}

	allowedCIRDs := []string{}
	if normaliseAllowedIP(ip) != "" {
		allowedCIRDs = append(allowedCIRDs, ip)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	existingRows, err := p.listPeerRows(ctx)
	if err != nil {
		p.log.Warnw("peer.create.duplicate_check_failed", "err", err)
	} else {
		if peerNameExists(existingRows, name) {
			p.renderPeersCreateError(c, http.StatusConflict, "Peer name already exists. Please use a unique name for this peer.")
			return
		}

		if allowedIPExists(existingRows, ip, "") {
			p.renderPeersCreateError(c, http.StatusConflict, "Allowed IP already exists. Please use a unique Allowed IP for this peer.")
			return
		}
	}

	req := &gatewayv2.CreatePeerRequest{
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
		p.renderPeersCreateError(c, peerErrorHTTPStatus(err), peerActionError("create", err))
		return
	}

	p.log.Infow("peer.create.success", "name", name, "pubkey", pubkey, "config_applied", resp.ConfigApplied)

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

	c.HTML(status, "peers.tmpl", peerPageData("Peers", rows, errorMsg))
}
