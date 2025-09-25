package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type peer struct {
	ID string
	Name string
	IP string
	Note string
}

type Peers struct {
	log *zap.SugaredLogger
	mu sync.Mutex
	data map[string]peer
	seq int
}

func NewPeers(log *zap.SugaredLogger) *Peers {
	return &Peers{
		log: log,
		data: map[string]peer{},
	}
}

func (p *Peers) List(c *gin.Context) {
	p.mu.Lock()
	rows := make([]peer, 0, len(p.data))
	for _, v := range p.data {
		rows = append(rows, v)
	}
	p.mu.Unlock()

	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })

	c.HTML(http.StatusOK, "peers.tmpl", gin.H{
		"title": "Peers",
		"peers": rows,
	})
}

func (p *Peers) Create(c *gin.Context) {
	name := c.PostForm("name")
	ip := c.PostForm("ip")
	note := c.PostForm("note")

	p.mu.Lock()
	p.seq++
	id := fmt.Sprintf("peer-%03d", p.seq)
	rec := peer{ID: id, Name: name, IP: ip, Note: note}
	p.data[id] = rec
	p.mu.Unlock()

	p.log.Infow("peer.create", "id", id, "name", name, "ip", ip, "note", note)
	// TODO: later call gRPC CreatePeer
	c.Redirect(http.StatusSeeOther, "/peers")
}

func (p *Peers) Delete(c *gin.Context) {
	id := c.Param("id")

	p.mu.Lock()
	delete(p.data, id)
	p.mu.Unlock()

	p.log.Infow("peer.delete", "id", id)
	// TODO: later call gRPC DeletePeer
	c.Redirect(http.StatusSeeOther, "/peers")
}
