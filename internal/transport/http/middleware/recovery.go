package middleware

import (
	"key-value-store/internal/util"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				util.WriteInternalError(w, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
