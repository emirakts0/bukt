package main

import (
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
	configs := config.Get()

	logger.Initialize(logger.Config{
		Environment: configs.Logging.Environment,
		LogLevel:    configs.Logging.Level,
	})

	slog.Info("Starting Lyko Key-Value Store.")
	slog.Info("Server starting...", "port", configs.Server.Port, "environment", configs.Logging.Environment, "log level", configs.Logging.Level)

	storageService := service.NewStorageService()
	kvHandler := handler.NewKVHandler(storageService)
	router := http.NewRouter(kvHandler)

	addr := ":" + strconv.Itoa(configs.Server.Port)
	slog.Info("Server started and listening", "address", addr)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
