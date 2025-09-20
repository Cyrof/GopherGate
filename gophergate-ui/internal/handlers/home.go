package handlers

import(
	"github.com/gin-gonic/gin"
)

func Home() gin.HandlerFunc{
	return func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{
			"title": "GopherGate UI",
			"message": "Hello from Gin (Zap logger wired).",
		})
	}
}