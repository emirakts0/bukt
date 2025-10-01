package middleware

import (
	"key-value-store/internal/util"
	"log/slog"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("PANIC", "errs", err)
				util.WriteInternalError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
