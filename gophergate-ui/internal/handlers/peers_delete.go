package handlers

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func (p *Peers) Delete(c *gin.Context) {
	pubkeyParam := c.PostForm("pubkey")

	p.mu.Lock()
	defer p.mu.Unlock()

	index := -1
	for i, v := range p.data {
		if v.PublicKey == pubkeyParam {
			index = i
			break
		}
	}

	if index == -1 {
		p.log.Warnw("peer.delete.missing", "pubkey", pubkeyParam)
		c.Redirect(http.StatusSeeOther, "/peers")
		return
	}

	p.data = slices.Delete(p.data, index, index+1)
	p.log.Infow("peer.delete", "pubkey", pubkeyParam)

	// TODO: later call gRPC DeletePeer
	c.Redirect(http.StatusSeeOther, "/peers")
}
