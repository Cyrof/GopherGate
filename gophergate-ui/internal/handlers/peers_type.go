package handlers

import (
	"sync"

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
