package main

import (
	"context"
	"fmt"
	"key-value-store/internal/auth"
	"key-value-store/internal/bucket"
	"key-value-store/internal/config"
	"key-value-store/internal/logger"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http"
	"key-value-store/internal/transport/tcp"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	configs := config.NewConfig()

	logger.Initialize(logger.Config{
		Environment: configs.Logging.Environment,
		LogLevel:    configs.Logging.Level,
	})

	slog.Info("Configuration loaded",
		slog.Group("server",
			slog.Int("http_port", configs.Server.Port),
			slog.Int("tcp_port", configs.Server.TCPPort),
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
	fmt.Println(auth.Manager().GenerateToken("default", 0))

	// Create bucket manager
	bucketManager := bucket.NewBucketManager(configs)

	// Create services
	storageService := service.NewStorageService(bucketManager, configs)
	bucketService := service.NewBucketService(bucketManager)

	// Create HTTP router
	httpRouter := http.NewRouter(storageService, bucketService)

	// Create TCP handler and server
	tcpHandler := tcp.NewHandler(storageService)
	tcpAddr := "tcp://:" + strconv.Itoa(configs.Server.TCPPort)
	tcpServer := tcp.NewServer(tcpAddr, tcpHandler)

	// Start TCP server in goroutine
	go func() {
		slog.Info("TCP Server starting", "address", tcpAddr)
		if err := tcpServer.Start(); err != nil {
			log.Fatal("Failed to start TCP server", err)
		}
	}()

	// Start HTTP server in goroutine
	httpAddr := ":" + strconv.Itoa(configs.Server.Port)
	go func() {
		slog.Info("HTTP Server starting", "address", httpAddr)
		if err := httpRouter.Run(httpAddr); err != nil {
			log.Fatal("Failed to start HTTP server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down servers...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop TCP server
	if err := tcpServer.Stop(ctx); err != nil {
		slog.Error("TCP Server shutdown error", "error", err)
	}

	slog.Info("Servers stopped gracefully")
}
