package middleware

import (
	"key-value-store/internal/util"
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		query := r.URL.RawQuery
		method := r.Method
		clientIP := r.RemoteAddr
		userAgent := r.UserAgent()
		correlationID := util.GetCorrelationID(r.Context())

		slog.Info("HTTP/REQUEST",
			slog.String("crr-id", correlationID),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", userAgent),
		)

		rw := &responseWriter{ResponseWriter: w, statusCode: 0}
		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		attrs := []any{
			slog.String("crr-id", correlationID),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", userAgent),
			slog.Int("status", rw.statusCode),
			slog.Duration("latency", latency),
		}

		switch {
		case rw.statusCode >= 500:
			slog.Error("HTTP/RESPONSE", attrs...)
		default:
			slog.Info("HTTP/RESPONSE", attrs...)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
