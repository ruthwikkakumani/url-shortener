package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Only log if the request actually failed (400s and 500s)
		if status >= 400 {
			logger.Error("request failed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.String("ip", c.ClientIP()),
				zap.Duration("latency", duration),
			)
		}

		/*
			logger.Info("incoming request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("ip", c.ClientIP()),
				zap.Duration("latency", duration),
				zap.String("user-agent", c.Request.UserAgent()),
			)
		*/
	}
}
