package middleware

import (
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
		correlationID := CorrelationID(r.Context())

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		latency := time.Since(start)

		attrs := []any{
			slog.String("crr-id", correlationID),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", userAgent),
			slog.Int64("latency_ns", latency.Nanoseconds()),
			slog.Int("status", rw.statusCode),
		}

		switch {
		case rw.statusCode >= 500:
			slog.Error("REQUEST", attrs...)
		case rw.statusCode >= 400:
			slog.Warn("REQUEST", attrs...)
		default:
			slog.Info("REQUEST", attrs...)
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
