package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
)

func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		correlationID := c.GetHeader(CorrelationIDHeader)

		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		c.Set("correlation_id", correlationID)

		c.Header(CorrelationIDHeader, correlationID)

		c.Next()
	}
}
