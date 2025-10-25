package handlers

import (
	"net/http"
	"slices"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type peer struct {
	Name      string
	IP        string
	Keepalive string
	PublicKey string
}

type Peers struct {
	log  *zap.SugaredLogger
	mu   sync.Mutex
	data []peer
}

func NewPeers(log *zap.SugaredLogger) *Peers {
	return &Peers{
		log:  log,
		data: []peer{},
	}
}

func (p *Peers) List(c *gin.Context) {
	p.mu.Lock()
	rows := make([]peer, 0, len(p.data))
	for _, v := range p.data {
		rows = append(rows, v)
	}
	p.mu.Unlock()

	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

	c.HTML(http.StatusOK, "peers.tmpl", gin.H{
		"title": "Peers",
		"peers": rows,
	})
}

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
