package handlers

import (
	"sync"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"go.uber.org/zap"
)

type peer struct {
	Name      string
	IP        string
	Keepalive string
	PublicKey string
	Endpoint  string
	RxBytes   uint64
	TxBytes   uint64
	Handshake string
}

type Peers struct {
	log  *zap.SugaredLogger
	mu   sync.Mutex
	data []peer
	grpc *grpcclient.Client
	wgIface string
}

func NewPeers(log *zap.SugaredLogger, grpcClient *grpcclient.Client, wgIface string) *Peers {
	return &Peers{
		log:  log,
		data: []peer{},
		grpc: grpcClient,
		wgIface: wgIface,
	}
}
