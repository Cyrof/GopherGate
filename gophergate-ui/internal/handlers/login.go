package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Cyrof/GopherGate/gophergate-ui/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db  *pgxpool.Pool
	log *zap.SugaredLogger
}

func NewAuthHandler(db *pgxpool.Pool, log *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{db: db, log: log}
}

// LoginPage renders the login form (GET /)
func (h *AuthHandler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "layouts/auth", gin.H{
			"title": "Login",
		})
	}
}

// LoginPost handles the login form submission (POST /login)
func (h *AuthHandler) LoginPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "" || password == "" {
			c.HTML(http.StatusOK, "layouts/auth", gin.H{
				"title": "Login",
				"error": "Username and password are required",
			})
			return
		}

		// look up user in db
		var (
			userID     int64
			hashedPass string
			role       string
		)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := h.db.QueryRow(ctx,
			"SELECT id, password, role FROM users WHERE username = $1",
			username,
		).Scan(&userID, &hashedPass, &role)

		if err != nil {
			h.log.Infow("login failed: user not found", "username", username)
			c.HTML(http.StatusOK, "layouts/auth", gin.H{
				"title": "Login",
				"error": "Invalid username or password",
			})
			return
		}

		// compare password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password)); err != nil {
			h.log.Infow("login failed: wrong password", "username", username)
			c.HTML(http.StatusOK, "layouts/auth", gin.H{
				"title": "Login",
				"error": "Invalid username or password",
			})
			return
		}

		// set session
		if err := auth.SetSession(c, userID, username, role); err != nil {
			h.log.Errorw("failed to set session", "err", err)
			c.HTML(http.StatusInternalServerError, "layouts/auth", gin.H{
				"title": "Login",
				"error": "Something went wrong. Please try again.",
			})
			return
		}

		h.log.Infow("login success", "username", username, "role", role)
		c.Redirect(http.StatusFound, "/dashboard")
	}
}

// logout handles the logout (POST /logout)
func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := auth.ClearSession(c); err != nil {
			h.log.Errorw("failed to clear session", "err", err)
		}
		c.Redirect(http.StatusFound, "/")
	}
}
