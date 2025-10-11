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

func NewRouter(kvHandler *handler.KVHandler, bucketHandler *handler.BucketHandler) *Router {
	mux := http.NewServeMux()

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
	mux.HandleFunc("POST /api/buckets", middleware.ApplyMiddleware(bucketHandler.CreateBucket, noAuthMw...))
	mux.HandleFunc("GET /api/buckets", middleware.ApplyMiddleware(bucketHandler.ListBuckets, mw...))
	mux.HandleFunc("GET /api/buckets/{bucket}", middleware.ApplyMiddleware(bucketHandler.GetBucket, mw...))
	mux.HandleFunc("DELETE /api/buckets/{bucket}", middleware.ApplyMiddleware(bucketHandler.DeleteBucket, mw...))

	// todo: add get random.
	// Key-value endpoints
	mux.HandleFunc("POST /api/buckets/{bucket}/kv", middleware.ApplyMiddleware(kvHandler.Create, mw...))
	mux.HandleFunc("GET /api/buckets/{bucket}/kv/{key}", middleware.ApplyMiddleware(kvHandler.Get, mw...))
	mux.HandleFunc("DELETE /api/buckets/{bucket}/kv/{key}", middleware.ApplyMiddleware(kvHandler.Delete, mw...))

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
