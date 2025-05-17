package middleware

import (
	"encoding/base64"
	"key-value-store/internal/config"
	"key-value-store/internal/util/http_util"
	"net/http"
	"strings"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http_util.WriteUnauthorized(w, "Authorization header is required")
			return
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			http_util.WriteUnauthorized(w, "Invalid authorization header format")
			return
		}

		// Decode the base64 credentials
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http_util.WriteUnauthorized(w, "Invalid credentials format")
			return
		}

		// Split username and password
		credentials := strings.Split(string(decoded), ":")
		if len(credentials) != 2 {
			http_util.WriteUnauthorized(w, "Invalid credentials format")

			return
		}

		// Validate credentials
		if credentials[0] != config.Config().Auth.Username || credentials[1] != config.Config().Auth.Password {
			http_util.WriteUnauthorized(w, "Invalid credentials")
			return
		}

		next.ServeHTTP(w, r)
	})
}
