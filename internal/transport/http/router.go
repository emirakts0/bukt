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

	commonMiddleware := []middleware.Middleware{
		middleware.Recovery,
		middleware.Correlation,
		middleware.Logger,
		middleware.Auth,
	}

	// Bucket management endpoints
	mux.HandleFunc("POST /api/buckets", middleware.ApplyMiddleware(bucketHandler.CreateBucket, commonMiddleware...))
	mux.HandleFunc("GET /api/buckets", middleware.ApplyMiddleware(bucketHandler.ListBuckets, commonMiddleware...))
	mux.HandleFunc("GET /api/buckets/{bucket}", middleware.ApplyMiddleware(bucketHandler.GetBucket, commonMiddleware...))
	mux.HandleFunc("DELETE /api/buckets/{bucket}", middleware.ApplyMiddleware(bucketHandler.DeleteBucket, commonMiddleware...))

	// todo: add get random.
	// Key-value endpoints
	mux.HandleFunc("POST /api/buckets/{bucket}/kv", middleware.ApplyMiddleware(kvHandler.Create, commonMiddleware...))
	mux.HandleFunc("GET /api/buckets/{bucket}/kv/{key}", middleware.ApplyMiddleware(kvHandler.Get, commonMiddleware...))
	mux.HandleFunc("DELETE /api/buckets/{bucket}/kv/{key}", middleware.ApplyMiddleware(kvHandler.Delete, commonMiddleware...))

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
