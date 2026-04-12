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
