package middleware

import (
	"encoding/base64"
	"key-value-store/internal/config"
	"key-value-store/internal/util/http_util"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	cfg config.AuthConfig
}

func NewAuthMiddleware(cfg config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http_util.WriteUnauthorized(w, "Authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			http_util.WriteUnauthorized(w, "Invalid authorization header format")
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http_util.WriteUnauthorized(w, "Invalid credentials format")
			return
		}

		credentials := strings.Split(string(decoded), ":")
		if len(credentials) != 2 {
			http_util.WriteUnauthorized(w, "Invalid credentials format")
			return
		}

		if credentials[0] != am.cfg.Username || credentials[1] != am.cfg.Password {
			http_util.WriteUnauthorized(w, "Invalid credentials")
			return
		}

		next.ServeHTTP(w, r)
	})
}
