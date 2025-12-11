package handlers

import (
	"context"
	"net/http"
	"sort"
	"time"

	gatewayv1 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v1"
	"github.com/gin-gonic/gin"
)

func (p *Peers) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := p.grpc.ListPeer(ctx, &gatewayv1.ListPeerRequest{
		Iface: p.wgIface.
	})

	if err != nil {
		p.log.Errorw("peer.list.grpc_error", "err", err)
		c.HTML(http.StatusInternalServerError, "peers.tmpl", gin.H{
			"title": "Peers",
			"peers": []peer{},
			"error": "Failed to fetch peers from backend",
		})
		return
	}

	row := make([]peer, 0, len(resp.Peers))
	for _, peerStatus := range resp.Peers {
		ip := ""
		if len(peerStatus.AllowedIps) > 0 {
			ip = peerStatus.AllowedIps[0]
		}

		rows = append(rows, peer{
			Name: "",	// Name not stored in proto hence need a separate storage
			IP: ip,
			Keepalive: peerStatus.Keepalive,
			PublicKey: peerStatus.PublicKey,
			Endpoint: peerStatus.Endpoint,
			RxBytes: peerStatus.RxBytes,
			TxBytes: peerStatus.TxBytes,
			Handshake: peerStatus.Handshake,
		})

	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].PublicKey < rows[j].PublicKey })

	c.HTML(http.StatusOK, "peers.tmpl", gin.H{
		"title": "Peers",
		"peers": rows,
	})
}
