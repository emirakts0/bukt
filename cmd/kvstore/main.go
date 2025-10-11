package main

import (
	"key-value-store/internal/auth"
	"key-value-store/internal/bucket"
	"key-value-store/internal/config"
	"key-value-store/internal/logger"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http"
	"key-value-store/internal/transport/http/handler"
	"log"
	"log/slog"
	"strconv"
)

func main() {
	configs := config.NewConfig()

	logger.Initialize(logger.Config{
		Environment: configs.Logging.Environment,
		LogLevel:    configs.Logging.Level,
	})

	slog.Info("Configuration loaded",
		slog.Group("server",
			slog.Int("port", configs.Server.Port),
		),
		slog.Group("store",
			slog.Int("shard_count", configs.Store.ShardCount),
		),
	)

	slog.Info("Starting Lyko Key-Value Store.")
	slog.Info("Server starting...", "port", configs.Server.Port, "environment", configs.Logging.Environment, "log level", configs.Logging.Level)

	// Initialize singleton auth manager with secret key
	auth.Initialize(configs.Auth.TokenSecret)
	slog.Info("Auth manager initialized")

	// Create bucket manager
	bucketManager := bucket.NewBucketManager(configs)

	// Create services
	storageService := service.NewStorageService(bucketManager, configs)
	bucketService := service.NewBucketService(bucketManager)

	// Create handlers
	kvHandler := handler.NewKVHandler(storageService)
	bucketHandler := handler.NewBucketHandler(bucketService)

	// Create router
	router := http.NewRouter(kvHandler, bucketHandler)

	addr := ":" + strconv.Itoa(configs.Server.Port)
	slog.Info("Server started and listening", "address", addr)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
