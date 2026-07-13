package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapLogger(l *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		l.Infow("http",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

func ZapRecovery(l *zap.SugaredLogger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rec interface{}) {
		l.Errorw("panic", "err", rec)
		c.AbortWithStatus(500)
	})
}
