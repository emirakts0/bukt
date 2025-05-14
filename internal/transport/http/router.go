package http

import (
	"key-value-store/internal/transport/http/handler"
	"key-value-store/internal/transport/http/middleware"
	"net/http"
)

type Router struct {
	server *http.Server
	mux    *http.ServeMux
}

func NewRouter(kvHandler *handler.KVHandler) *Router {
	mux := http.NewServeMux()

	commonMiddleware := []middleware.Middleware{
		middleware.Recovery,
		middleware.Logger,
		//middleware.Correlation,
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
		Addr:    addr,
		Handler: r.mux,
	}
	return r.server.ListenAndServe()
}
