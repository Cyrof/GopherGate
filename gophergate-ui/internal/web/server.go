package web

// this file should hold the initialisation function for the httpserver using gin as the framework

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/auth"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/config"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/grpcclient"
	"github.com/Cyrof/GopherGate/gophergate-ui/internal/handlers"
	assetweb "github.com/Cyrof/GopherGate/gophergate-ui/web"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func Run(cfg *config.Config, log *zap.SugaredLogger, db *pgxpool.Pool) (err error) {
	defer recoverRunPanic(log, &err)

	log.Infow("starting web server", "http", cfg.HTTPAddr, "grpc", cfg.GRPCAddr)
	// Initialised gRPC client
	grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return fmt.Errorf("init grpc client: %w", err)
	}
	defer func() { _ = grpcClient.Close() }()

	router := gin.New()
	router.Use(ZapLogger(log), ZapRecovery(log))

	// setup session middleware
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   false,
	})
	router.Use(sessions.Sessions("gophergate-session", store))

	log.Infow("mounting static files...")
	if err := mountStatic(router); err != nil {
		return fmt.Errorf("mount static files: %w", err)
	}

	log.Infow("loading templates...")
	if err := loadTemplates(router); err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	log.Infow("registering routes...")
	if err := registerRoutes(router, cfg, log, *grpcClient, db); err != nil {
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
	db *pgxpool.Pool,
) error {
	authHandler := handlers.NewAuthHandler(db, log)
	router.GET("/", authHandler.LoginPage())
	router.POST("/login", authHandler.LoginPost())
	router.POST("/logout", authHandler.Logout())
	router.GET("/logout", authHandler.Logout()) // temp for dev
	log.Infow("auth route registered")

	// health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	log.Infow("health route registerd")

	// Protected route (auth required)
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		dashboardHandler := handlers.Dashboard()
		if dashboardHandler == nil {
			return fmt.Errorf("handlers.Dashboard() returned nil")
		}
		protected.GET("/dashboard", dashboardHandler)
		log.Infow("dashboard route registered")

		peer := handlers.NewPeers(log, &grpcClient, cfg.WGIface)
		if peer == nil {
			return fmt.Errorf("handlers.NewPeers() returned nil")
		}
		protected.GET("/peers", peer.List)
		protected.POST("/peers", peer.Create)
		protected.GET("/peers/edit", peer.EditForm)
		protected.POST("/peers/edit", peer.Edit)
		protected.POST("/peers/delete", peer.Delete)
		log.Infow("peer routes registered")
	}

	return nil
}
