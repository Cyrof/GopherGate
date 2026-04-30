package handlers

import (
	"context"
	"net/http"
	"time"

	gatewayv2 "github.com/Cyrof/GopherGate/gophergate-core/pkg/gen/gateway/v2"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"github.com/gin-gonic/gin"
)

func Dashboard(grpcClient *grpcclient.Client, iface string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := grpcClient.GetDashboard(ctx, &gatewayv2.GetDashboardRequest{
			Iface: iface,
		})
		if err != nil {
			c.HTML(http.StatusBadGateway, "pages/index.tmpl", gin.H{
				"title":      "Dashboard",
				"pagetitle":  "Dashboard",
				"activeNav":  "dashboard",
				"statusText": "All Systems Operational",
				"error":      err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "pages/index.tmpl", gin.H{
			"title":      "Dashboard",
			"pageTitle":  "Dashboard",
			"activeNav":  "dashboard",
			"statusText": "All Systems Operational",
			"summary":    resp.GetSummary(),
			"notes":      resp.GetNotes(),
			"alerts":     resp.GetAlerts(),
			"peers":      resp.GetPeers(),
		})
	}
}


func DashboardModal(grpcClient *grpcclient.Client, iface string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		publicKey := c.Param("publicKey")
		if publicKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing public key",
			})
			return
		}

		rangeValue := c.DefaultQuery("range", "24h")

		resp, err := grpcClient.GetPeerTraffic(ctx, &gatewayv2.GetPeerTrafficRequest{
			Iface: iface,
			PublicKey: publicKey,
			Range:     rangeValue,
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "failed to retrieve peer traffic",
			})
			return
		}

		dashboardResp, err := grpcClient.GetDashboard(ctx, &gatewayv2.GetDashboardRequest{
			Iface: iface,
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "failed to retrieve dashboard peer overview",
			})
			return
		}

		traffic := resp.GetTraffic()
		if traffic == nil {
			c.JSON(http.StatusOK, gin.H{
				"name":       "",
				"public_key": publicKey,
				"points":     []any{},
			})
			return
		}

		var allowedIPs []string
		for _, peer := range dashboardResp.GetPeers() {
			if peer.GetPublicKey() == publicKey {
				allowedIPs = peer.GetAllowedIps()
				break
			}
		}

		var totalRx uint64
		var totalTx uint64

		points := make([]gin.H, 0, len(traffic.GetPoints()))
		for _, p := range traffic.GetPoints() {
			totalRx = p.GetRxBytes()
			totalTx = p.GetTxBytes()

			points = append(points, gin.H{
				"timestamp":   p.GetTimestamp(),
				"rx_bytes":    p.GetRxBytes(),
				"tx_bytes":    p.GetTxBytes(),
				"total_bytes": p.GetTotalBytes(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"name":       traffic.GetName(),
			"public_key": traffic.GetPublicKey(),
			"allowed_ips": allowedIPs,
			"total_rx":   totalRx,
			"total_tx":   totalTx,
			"points":     points,
		})
	}
}