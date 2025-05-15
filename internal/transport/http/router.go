package http

import (
	"key-value-store/internal/transport/http/handler"
	"key-value-store/internal/transport/http/middleware"
	"net/http"
	"time"
)

const (
	readTimeout  = 5  // in seconds
	writeTimeout = 10 // in seconds
	idleTimeout  = 15 // in seconds
)

type Router struct {
	server *http.Server
	mux    *http.ServeMux
}

func NewRouter(kvHandler *handler.KVHandler) *Router {
	mux := http.NewServeMux()

	commonMiddleware := []middleware.Middleware{
		middleware.Recovery,
		middleware.Correlation,
		middleware.Logger,
		middleware.Auth,
	}

	mux.HandleFunc("POST /api/kv", middleware.ApplyMiddleware(kvHandler.Create, commonMiddleware...))
	mux.HandleFunc("GET /api/kv/{key}", middleware.ApplyMiddleware(kvHandler.Get, commonMiddleware...))
	mux.HandleFunc("DELETE /api/kv/{key}", middleware.ApplyMiddleware(kvHandler.Delete, commonMiddleware...))

	return &Router{
		mux: mux,
	}
}

func (r *Router) Run(addr string) error {
	r.server = &http.Server{
		Addr:         addr,
		Handler:      r.mux,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}
	return r.server.ListenAndServe()
}
