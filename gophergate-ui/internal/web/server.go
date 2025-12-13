package web

// this file should hold the initialisation function for the httpserver using gin as the framework

import (
	"net/http"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(cfg *config.Config, log *zap.SugaredLogger) error {
	// Initialised gRPC client
	grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return err
	}
	defer grpcClient.Close()

	r := gin.New()
	r.Use(ZapLogger(log), ZapRecovery(log))

	r.StaticFS("/static", http.FS(webFS))
	r.LoadHTMLFS(webFS, "web/templates/*.tmpl")

	r.GET("/", handlers.Home())

	peer := handlers.NewPeers(log, grpcClient, cfg.WGIface)
	r.GET("/peers", peer.List)
	r.POST("/peers", peer.Create)

	r.GET("/peers/edit", peer.EditForm)
	r.POST("/peers/edit", peer.Edit)

	r.POST("/peers/delete", peer.Delete)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r.Run(cfg.HTTPAddr)
}

