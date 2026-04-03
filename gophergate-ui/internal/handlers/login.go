package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "layouts/auth", gin.H{
			"title":   "Login",
			"content": "pages/auth/login",
		})
	}
}
