package handlers

import (
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
		data: map[string]peer{
			"peer-001": {ID: "peer-001", Name:"alice", IP: "10.8.0.2/32", Note: "seed"},
			"peer-002": {ID: "peer-002", Name:"bob", IP: "10.8.0.3/32", Note: "seed"},
		},
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
