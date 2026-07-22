package handlers

import (
	"context"
	"net/http"
	"time"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/gin-gonic/gin"
)

func (p *Peers) Edit(c *gin.Context) {
	pubkeyOriginal := c.PostForm("pubkey_original")
	pubkey := c.PostForm("pubkey")
	ip := c.PostForm("ip")
	keepaliveStr := c.PostForm("keepalive")
	endpoint := normaliseEndpoint(c.PostForm("endpoint"))

	rec := peer{
		PublicKey: pubkey,
		IP:        ip,
		Keepalive: keepaliveStr,
		Endpoint:  endpoint,
	}

	keepalive, err := parseKeepaliveSeconds(keepaliveStr)
	if err != nil {
		p.log.Warnw("peer.update.invalid_keepalive", "value", keepaliveStr, "err", err)
		p.renderPeerEditPage(c, http.StatusBadRequest, rec, "Invalid keepalive value. Please enter a number of seconds, for example 25.")
		return
	}

	if err := validateAllowedCIDR(ip); err != nil {
		p.log.Warnw("peer.update.invalid_allowed_ip", "value", ip, "err", err)
		p.renderPeerEditPage(c, http.StatusBadRequest, rec, "Invalid Allowed IP. Please enter a host IP or CIDR, for example 10.13.13.2 or 10.13.13.2/32.")
		return
	}

	if normaliseAllowedIP(ip) != "" {
		addr, _ := parsePeerIPInput(ip)
		ip = hostCIDR(addr)
		rec.IP = ip
	}

	if err := validateEndpoint(endpoint); err != nil {
		p.log.Warnw("peer.update.invalid_endpoint", "value", endpoint, "err", err)
		p.renderPeerEditPage(c, http.StatusBadRequest, rec, "Invalid endpoint. Leave it blank for roaming peers, or use the format IP:port, for example 192.168.1.100:51820")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	existingRows, err := p.listPeerRows(ctx)
	if err != nil {
		p.log.Warnw("peer.edit.duplicate_check_failed", "err", err)
	} else if allowedIPExists(existingRows, ip, pubkeyOriginal) {
		p.renderPeerEditPage(c, http.StatusConflict, rec, "Allowed IP already exists. Please use a unique Allowed IP for this peer.")
		return
	}

	req := &gatewayv2.UpdatePeerRequest{
		Iface:            p.wgIface,
		PublicKey:        pubkeyOriginal,
		Endpoint:         endpoint,
		KeepaliveSeconds: keepalive,
	}

	if normaliseAllowedIP(ip) != "" {
		req.SetAllowedCidrs = []string{ip}
	}

	resp, err := p.grpc.UpdatePeer(ctx, req)
	if err != nil {
		p.log.Errorw("peer.update.grpc_error", "pubkey", pubkeyOriginal, "err", err)
		p.renderPeerEditPage(c, peerErrorHTTPStatus(err), rec, peerActionError("update", err))
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
