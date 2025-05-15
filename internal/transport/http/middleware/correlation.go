package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	correlationIDKey    = "correlation_id"
)

func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationIDHeader)

		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)

		w.Header().Set(CorrelationIDHeader, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}
