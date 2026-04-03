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
	assetweb "github.com/Cyrof/GopherGate/gophergate-ui/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(cfg *config.Config, log *zap.SugaredLogger) (err error) {
	defer recoverRunPanic(log, &err)

	log.Infow("starting web server", "http", cfg.HTTPAddr, "grpc", cfg.GRPCAddr)
	// Initialised gRPC client
	grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return fmt.Errorf("init grpc client: %w", err)
	}
	defer grpcClient.Close()

	router := gin.New()
	router.Use(ZapLogger(log), ZapRecovery(log))

	log.Infow("mounting static files...")
	if err := mountStatic(router); err != nil {
		return fmt.Errorf("mount static files: %w", err)
	}

	log.Infow("loading templates...")
	if err := loadTemplates(router); err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	log.Infow("registering routes...")
	if err := registerRoutes(router, cfg, log, *grpcClient); err != nil {
		return fmt.Errorf("register routes: %w", err)
	}

	log.Infow("server starting", "addr", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Errorw("gin server run failed", "error", err)
		return fmt.Errorf("run gin server: %w", err)
	}

	return nil
}

func recoverRunPanic(log *zap.SugaredLogger, errp *error) {
	if r := recover(); r != nil {
		recErr := fmt.Errorf("panic recovered: %v", r)
		log.Errorw("panic recovered in web.Run", "err", recErr)
		*errp = recErr
	}
}

func mountStatic(router *gin.Engine) error {
	staticFS, err := fs.Sub(assetweb.FS, "static")
	if err != nil {
		return err
	}
	router.StaticFS("/static", http.FS(staticFS))
	return nil
}

func loadTemplates(router *gin.Engine) error {
	if _, err := fs.Stat(assetweb.FS, "templates"); err != nil {
		return err
	}
	tmpl, err := template.ParseFS(
		assetweb.FS,
		"templates/layouts/*.tmpl",
		"templates/partials/*.tmpl",
		"templates/pages/*.tmpl",
		"templates/pages/peers/*.tmpl",
		"templates/pages/auth/*.tmpl",
	)
	if err != nil {
		return err
	}

	router.SetHTMLTemplate(tmpl)
	return nil
}

func registerRoutes(
	router *gin.Engine,
	cfg *config.Config,
	log *zap.SugaredLogger,
	grpcClient grpcclient.Client,
) error {
	loginHandler := handlers.Login()
	if loginHandler == nil {
		return fmt.Errorf("handlers.Login() returned nil")
	}
	router.GET("/", loginHandler)
	log.Infow("login route registered")

	dashboardHandler := handlers.Dashboard()
	if dashboardHandler == nil {
		return fmt.Errorf("handlers.Dashboard() returned nil")
	}
	router.GET("/dashboard", dashboardHandler)
	log.Infow("dashboard router registered")

	peer := handlers.NewPeers(log, &grpcClient, cfg.WGIface)
	if peer == nil {
		return fmt.Errorf("handlers.NewPeers() returned nil")
	}

	router.GET("/peers", peer.List)
	router.POST("/peers", peer.Create)
	router.GET("/peers/edit", peer.EditForm)
	router.POST("/peers/edit", peer.Edit)
	router.POST("/peers/delete", peer.Delete)
	log.Infow("peer routes registed")

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	log.Infow("health route registered")

	return nil
}
