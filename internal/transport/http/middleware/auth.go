package middleware

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"key-value-store/internal/config"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Decode the base64 credentials
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials format",
			})
			return
		}

		// Split username and password
		credentials := strings.Split(string(decoded), ":")
		if len(credentials) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials format",
			})
			return
		}

		// Validate credentials
		if credentials[0] != config.Get().Auth.Username || credentials[1] != config.Get().Auth.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			return
		}

		c.Next()
	}
}
