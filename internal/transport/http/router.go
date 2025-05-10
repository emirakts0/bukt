package http

import (
	"github.com/gin-gonic/gin"
	"key-value-store/internal/transport/http/handler"
	"key-value-store/internal/transport/http/middleware"
)

func NewRouter(kvHandler *handler.KVHandler) *gin.Engine {
	// Disable Gin's default logger
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()

	// Add our custom logger middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CorrelationMiddleware())
	router.Use(middleware.AuthMiddleware())

	// Add recovery middleware
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		kv := api.Group("/kv")
		{
			kv.POST("", kvHandler.Create)
			kv.GET("/:key", kvHandler.Get)
			kv.DELETE("/:key", kvHandler.Delete)
		}
	}

	return router
}
