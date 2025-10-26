package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *Peers) Edit(c *gin.Context) {
	pubkeyOriginal := c.PostForm("pubkey_original")

	name := c.PostForm("name")
	ip := c.PostForm("ip")
	keepalive := c.PostForm("keepalive")
	pubkey := c.PostForm("pubkey")

	p.mu.Lock()
	defer p.mu.Unlock()

	found := false
	for i := range p.data {
		if p.data[i].PublicKey == pubkeyOriginal {
			p.data[i].Name = name
			p.data[i].IP = ip
			p.data[i].Keepalive = keepalive
			p.data[i].PublicKey = pubkey
			found = true
			break
		}
	}

	if found {
		p.log.Infow("peer.update", "pubkey", pubkeyOriginal, "name", name, "ip", ip)
	} else {
		p.log.Warnw("peer.update.missing", "pubkey", pubkeyOriginal)
	}

	// TODO: later call gRPC UpdatePeer
	c.Redirect(http.StatusSeeOther, "/peers")
}
