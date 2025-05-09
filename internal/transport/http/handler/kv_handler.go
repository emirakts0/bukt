package handler

import (
	_ "context"
	"errors"
	"github.com/gin-gonic/gin"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"net/http"
)

type KVHandler struct {
	service *service.StorageService
}

func NewKVHandler(service *service.StorageService) *KVHandler {
	return &KVHandler{service: service}
}

func (h *KVHandler) Create(c *gin.Context) {
	var req request.CreateKVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	entry, err := h.service.Set(c.Request.Context(), req.Key, req.Value, req.TTL)

	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.KVResponse{
		Message:   "Key-value pair stored successfully",
		Key:       entry.Key,
		ExpiresAt: entry.ExpiresAt,
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *KVHandler) Get(c *gin.Context) {
	key := c.Param("key")

	entry, err := h.service.Get(c.Request.Context(), key)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, service.ErrKeyExpired) {
			status = http.StatusGone
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.KVResponse{
		Message:   "Key found",
		Key:       entry.Key,
		Value:     entry.Value,
		ExpiresAt: entry.ExpiresAt,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *KVHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	if err := h.service.Delete(c.Request.Context(), key); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, service.ErrKeyExpired) {
			status = http.StatusGone
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.KVResponse{
		Message: "Key deleted successfully",
		Key:     key,
	})
}
