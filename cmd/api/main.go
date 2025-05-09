package main

import (
	"fmt"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http"
	"key-value-store/internal/transport/http/handler"
	"log"
)

func main() {
	fmt.Println("Starting Lyko Key-Value Store...")

	storageService := service.NewStorageService()
	kvHandler := handler.NewKVHandler(storageService)
	router := http.NewRouter(kvHandler)

	// Start server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
