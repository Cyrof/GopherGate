package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	grpcClient, err := grpcclient.New(cfg.GRPCAddr, cfg.TLS, log)
	if err != nil {
		return fmt.Errorf("init grpc client: %w", err)
	}
	defer func() { _ = grpcClient.Close() }()

	router := gin.New()
	router.Use(ZapLogger(log), ZapRecovery(log))

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

	funcMap := template.FuncMap{
		"humanBytes":       humanBytes,
		"formatClock":      formatClock,
		"timeAgo":          timeAgo,
		"handshakeDisplay": handshakeDisplay,
		"stateDisplay":     stateDisplay,
		"statePillClass":   statePillClass,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(
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
	router.GET("/logout", authHandler.Logout())
	log.Infow("auth route registered")

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	log.Infow("health route registerd")

	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		dashboardHandler := handlers.Dashboard(&grpcClient, cfg.WGIface)
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

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := uint64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}

	value := float64(n) / float64(div)
	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}
	if exp >= len(suffixes) {
		exp = len(suffixes) - 1
	}

	if value >= 100 {
		return fmt.Sprintf("%.0f %s", value, suffixes[exp])
	}
	if value >= 10 {
		return fmt.Sprintf("%.1f %s", value, suffixes[exp])
	}
	return fmt.Sprintf("%.2f %s", value, suffixes[exp])
}

func formatClock(s string) string {
	if s == "" {
		return "—"
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "—"
	}

	return t.Local().Format("15:04:05")
}

func timeAgo(s string) string {
	if s == "" {
		return "—"
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "—"
	}

	d := time.Since(t)
	if d < 0 {
		d = 0
	}

	switch {
	case d < time.Minute:
		secs := int(d.Seconds())
		if secs <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%ds ago", secs)
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hrs := int(d.Hours())
		return fmt.Sprintf("%dh ago", hrs)
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

func handshakeDisplay(s string) string {
	if s == "" {
		return "—"
	}
	return timeAgo(s)
}

func stateDisplay(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "CONNECTED":
		return "Connected"
	case "INTERMITTENT":
		return "Intermittent"
	case "DOWN":
		return "Down"
	default:
		if s == "" {
			return "Unknown"
		}
		return strings.Title(strings.ToLower(s))
	}
}

func statePillClass(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "CONNECTED":
		return "border-success/40 bg-success/10 text-success"
	case "INTERMITTENT":
		return "border-warning/40 bg-warning/10 text-warning"
	case "DOWN":
		return "border-error/40 bg-error/10 text-error"
	default:
		return "border-base-300 bg-base-200/70 text-base-content/80"
	}
}

// Optional helper if you later want counts or formatted ints in templates.
func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}
