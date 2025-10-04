package main

import (
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
		slog.Group("auth",
			slog.String("username", configs.Auth.Username),
		),
		slog.Group("store",
			slog.Int("shard_count", configs.Store.ShardCount),
			slog.String("compression_type", configs.Store.CompressionType),
			slog.Int64("compression_threshold", configs.Store.CompressionThreshold),
		),
	)

	slog.Info("Starting Lyko Key-Value Store.")
	slog.Info("Server starting...", "port", configs.Server.Port, "environment", configs.Logging.Environment, "log level", configs.Logging.Level)

	bucketManager := bucket.NewBucketManager(configs)

	storageService := service.NewStorageService(bucketManager, configs)
	bucketService := service.NewBucketService(bucketManager)

	kvHandler := handler.NewKVHandler(storageService)
	bucketHandler := handler.NewBucketHandler(bucketService)

	router := http.NewRouter(kvHandler, bucketHandler, configs.Auth)

	addr := ":" + strconv.Itoa(configs.Server.Port)
	slog.Info("Server started and listening", "address", addr)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
