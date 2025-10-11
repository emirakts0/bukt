package middleware

import (
	"key-value-store/internal/auth"
	"key-value-store/internal/util"
	"log/slog"
	"net/http"
)

// Auth validates bucket tokens and adds bucket info to context
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crrid := util.GetCorrelationID(r.Context())

		// Get bucket name from path
		bucketName := r.PathValue("bucket")
		if bucketName == "" {
			slog.Debug("Middleware: Bucket name not in path", "crr-id", crrid)
			util.WriteBadRequest(w, "Bucket name is required")
			return
		}

		// Get token from header
		tokenStr := r.Header.Get("X-Bucket-Token")
		if tokenStr == "" {
			slog.Debug("Middleware: Bucket token not provided", "crr-id", crrid)
			util.WriteUnauthorized(w, "X-Bucket-Token header is required")
			return
		}

		// Validate token with bucket name
		if !auth.Manager().ValidateToken(tokenStr, bucketName) {
			slog.Debug("Middleware: Invalid bucket token", "crr-id", crrid, "bucket", bucketName)
			util.WriteUnauthorized(w, "Invalid bucket token")
			return
		}

		// Add bucket name to context
		ctx := util.SetBucketName(r.Context(), bucketName)

		slog.Debug("Middleware: Bucket authenticated", "crr-id", crrid, "bucket", bucketName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
