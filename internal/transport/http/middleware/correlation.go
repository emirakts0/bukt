package middleware

import (
	"key-value-store/internal/util"
	"net/http"

	"github.com/google/uuid"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
)

func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationIDHeader)

		if correlationID == "" {
			uuidV7, _ := uuid.NewV7()
			correlationID = uuidV7.String()
		}

		ctx := util.SetCorrelationID(r.Context(), correlationID)

		w.Header().Set(CorrelationIDHeader, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
