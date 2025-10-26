package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *Peers) EditForm(c *gin.Context) {
	pubkey := c.Query("pubkey")

	p.mu.Lock()
	var rec *peer
	for i := range p.data {
		if p.data[i].PublicKey == pubkey {
			rec = &p.data[i]
			break
		}
	}
	p.mu.Unlock()

	if rec == nil {
		c.String(http.StatusNotFound, "peer not found")
		return
	}

	c.HTML(http.StatusOK, "peers_edit.tmpl", gin.H{
		"title": "Edit Peer",
		"peer":  rec,
	})

}
