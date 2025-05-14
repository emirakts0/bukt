package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func ApplyMiddleware(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	var handler http.Handler = h

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
