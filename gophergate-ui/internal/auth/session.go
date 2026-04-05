package auth

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionUserID   = "user_id"
	SessionUsername = "username"
	SessionRole     = "role"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(SessionUserID)
		if userID == nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Set("username", session.Get(SessionUserID))
		c.Set("role", session.Get(SessionRole))
		c.Next()
	}
}

func SetSession(c *gin.Context, userID int64, username, role string) error {
	session := sessions.Default(c)
	session.Set(SessionUserID, userID)
	session.Set(SessionUsername, userID)
	session.Set(SessionRole, role)
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	})
	return session.Save()
}

func ClearSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: -1,
	})
	return session.Save()
}
