package http

import (
	"key-value-store/internal/service"
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

func NewRouter(storageService service.IStorageService, bucketService service.IBucketService) *Router {
	mux := http.NewServeMux()
	handlers := NewHandlers(storageService, bucketService)

	mw := []middleware.Middleware{
		middleware.Recovery,
		middleware.Correlation,
		middleware.Logger,
		middleware.Auth,
	}

	noAuthMw := []middleware.Middleware{
		middleware.Recovery,
		middleware.Correlation,
		middleware.Logger,
	}

	// Bucket management endpoints
	mux.HandleFunc("POST /api/buckets", middleware.ApplyMiddleware(handlers.CreateBucket, noAuthMw...))
	mux.HandleFunc("GET /api/buckets", middleware.ApplyMiddleware(handlers.ListBuckets, mw...))
	mux.HandleFunc("GET /api/buckets/{bucket}", middleware.ApplyMiddleware(handlers.GetBucket, mw...))
	mux.HandleFunc("DELETE /api/buckets/{bucket}", middleware.ApplyMiddleware(handlers.DeleteBucket, mw...))

	// Key-value endpoints
	mux.HandleFunc("POST /api/{bucket}/kv", middleware.ApplyMiddleware(handlers.CreateKV, mw...))
	mux.HandleFunc("GET /api/{bucket}/kv/{key}", middleware.ApplyMiddleware(handlers.GetKV, mw...))
	mux.HandleFunc("DELETE /api/{bucket}/kv/{key}", middleware.ApplyMiddleware(handlers.DeleteKV, mw...))

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
