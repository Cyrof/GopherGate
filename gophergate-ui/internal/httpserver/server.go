package httpserver

// this file should hold the initialisation function for the httpserver using gin as the framework

import (
	"net/http"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/handlers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(cfg *config.Config, log *zap.SugaredLogger) error {
	r := gin.New()
	r.Use(ZapLogger(log), ZapRecovery(log))

	r.Static("/static", "web/static")
	r.LoadHTMLGlob("web/templates/*.tmpl")

	r.GET("/", handlers.Home())

	peer := handlers.NewPeers(log)
	r.GET("/peers", peer.List)
	r.POST("/peers", peer.Create)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r.Run(cfg.HTTPAddr)
}