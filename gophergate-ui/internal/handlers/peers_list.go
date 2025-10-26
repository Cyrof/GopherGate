package handlers

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

func (p *Peers) List(c *gin.Context) {
	p.mu.Lock()
	rows := make([]peer, 0, len(p.data))

	func() {
		defer func() {
			if r := recover(); r != nil {
				p.log.Errorw("peer.list.iteration_panic", "panic", r)
			}
		}()

		for i := range p.data {
			v := p.data[i]

			if v.Name == "" || v.PublicKey == "" {
				p.log.Warnw("peer.list.skip_invalid", "index", i)
				continue
			}
			rows = append(rows, v)
		}
	}()

	p.mu.Unlock()

	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

	c.HTML(http.StatusOK, "peers.tmpl", gin.H{
		"title": "Peers",
		"peers": rows,
	})
}
