package main

import (
	"fmt"
	"go.uber.org/zap"
	"key-value-store/internal/config"
	"key-value-store/internal/logger"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http"
	"key-value-store/internal/transport/http/handler"
	"strconv"
)

func main() {
	configs := config.Get()

	// Initialize logger
	logger.Initialize(logger.Config{
		Environment: configs.Logging.Environment,
		LogLevel:    configs.Logging.Level,
	})
	log := logger.Get()
	defer logger.Sync()

	log.Info("Starting Lyko Key-Value Store...",
		zap.Int("port", configs.Server.Port),
		zap.String("environment", configs.Logging.Environment),
	)

	storageService := service.NewStorageService()
	kvHandler := handler.NewKVHandler(storageService)
	router := http.NewRouter(kvHandler)

	// Start server
	log.Info("Server is starting",
		zap.String("address", fmt.Sprintf(":%d", configs.Server.Port)),
	)
	if err := router.Run(":" + strconv.Itoa(configs.Server.Port)); err != nil {
		log.Fatal("Failed to start server",
			zap.Error(err),
		)
	}
}
