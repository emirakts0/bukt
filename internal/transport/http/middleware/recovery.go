package middleware

import (
	"key-value-store/internal/util/http_util"
	"log/slog"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("PANIC", "errs", err)
				http_util.WriteInternalError(w, "Internal server error.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
