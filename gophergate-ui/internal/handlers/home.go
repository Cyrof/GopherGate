package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Home() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "layouts/base", gin.H{
			"title":      "Home",
			"pageTitle":  "Home",
			"activePage": "dashboard",
		})
	}
}
