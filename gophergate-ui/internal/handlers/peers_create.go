package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *Peers) Create(c *gin.Context) {
	name := c.PostForm("name")
	ip := c.PostForm("ip")
	keepalive := c.PostForm("keepalive")
	pubkey := c.PostForm("pubkey")

	newPeer := peer{
		Name:      name,
		IP:        ip,
		Keepalive: keepalive,
		PublicKey: pubkey,
	}

	p.mu.Lock()
	p.data = append(p.data, newPeer)
	p.mu.Unlock()

	p.log.Infow("peer.create", "name", name, "ip", ip, "keepalive(s)", keepalive, "public-key", pubkey)
	// TODO: later call gRPC CreatePeer
	c.Redirect(http.StatusSeeOther, "/peers")
}
