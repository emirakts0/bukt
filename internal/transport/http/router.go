package http

import (
	"github.com/gin-gonic/gin"
	"key-value-store/internal/transport/http/handler"
)

func NewRouter(kvHandler *handler.KVHandler) *gin.Engine {
	router := gin.Default()

	setupRoutes(router, kvHandler)

	return router
}

func setupRoutes(router *gin.Engine, kvHandler *handler.KVHandler) *gin.RouterGroup {
	api := router.Group("/api")
	{
		kv := api.Group("/kv")
		{
			kv.POST("", kvHandler.Create)
			kv.GET("/:key", kvHandler.Get)
			kv.DELETE("/:key", kvHandler.Delete)
		}
	}
	return api
}
