package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "layouts/base", gin.H{
			"title":      "Dashboard",
			"pageTitle":  "Dashboard",
			"activePage": "dashboard",
		})
	}
}
