package web

// this file should hold the initialisation function for the httpserver using gin as the framework

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/handlers"
	"github.com/Cyrof/GopherGate/gophergate-ui/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(cfg *config.Config, log *zap.SugaredLogger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			fmt.Printf("PANIC: %v\n", r)
		}
	}()

	log.Infow("starting web.Run")
	// Initialised gRPC client
	grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return err
	}
	defer grpcClient.Close()
	log.Infow("grpc client initialised")

	r := gin.New()
	r.Use(ZapLogger(log), ZapRecovery(log))
	log.Infow("gin router initialised")

	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		log.Errorw("failed to get static sud FS", "error", err)
		fmt.Printf("ERROR: failed to get static sub FS: %v\n", err)
		return err
	}
	r.StaticFS("/static", http.FS(staticFS))
	log.Infow("static FS mounted")

	if _, err := fs.Stat(web.FS, "templates"); err != nil {
		log.Errorw("templates directory not found", "error", err)
		fmt.Printf("ERROR: templates directory not found: %v\n", err)
		return err
	}

	// tmpl, err := template.ParseFS(web.FS, "templates/*.tmpl")
	tmpl, err := template.ParseFS(
		web.FS,
		"templates/layouts/*.tmpl",
		"templates/partials/*.tmpl",
		"templates/pages/*.tmpl",
		"templates/pages/peers/*.tmpl",
	)
	if err != nil {
		log.Errorw("failed to parse templates", "error", err)
		fmt.Printf("ERROR: failed to parse templates: %v\n", err)
		return err
	}
	r.SetHTMLTemplate(tmpl)
	log.Infow("templates loaded")

	homeHandler := handlers.Home()
	if homeHandler == nil {
		err := fmt.Errorf("handlers.Home() return nil")
		log.Errorw("handler error", "error", err)
		fmt.Printf("ERROR: %v\n", err)
		return err
	}
	r.GET("/", homeHandler)
	log.Infow("home route registered")

	peer := handlers.NewPeers(log, grpcClient, cfg.WGIface)
	if peer == nil {
		err := fmt.Errorf("handlers.NewPeers() returned nil")
		log.Errorw("handler error", "error", err)
		fmt.Printf("ERROR: %v\n", err)
		return err
	}
	r.GET("/peers", peer.List)
	r.POST("/peers", peer.Create)
	r.GET("/peers/edit", peer.EditForm)
	r.POST("/peers/edit", peer.Edit)
	r.POST("/peers/delete", peer.Delete)
	log.Infow("peer routes registerd")

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	log.Infow("health route registered")

	log.Infow("starting server", "addr", cfg.HTTPAddr)

	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Errorw("server failed to start", "error", err)
		fmt.Printf("ERROR: server failed to start: %v\n", err)
		return err
	}

	return nil
}
