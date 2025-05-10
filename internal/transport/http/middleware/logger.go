package middleware

import (
	"key-value-store/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a gin middleware that logs requests using zap
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		log := logger.Get()

		switch {
		case statusCode >= 500:
			log.Error("Server error", fields...)
		case statusCode >= 400:
			log.Warn("Client error", fields...)
		default:
			log.Info("Request processed", fields...)
		}
	}
}
