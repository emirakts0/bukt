package handler

import (
	_ "context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"key-value-store/internal/errs"
	"key-value-store/internal/logger"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"net/http"
)

type KVHandler struct {
	service service.StorageService
	log     *zap.SugaredLogger
}

func NewKVHandler(service service.StorageService) *KVHandler {
	return &KVHandler{
		service: service,
		log:     logger.GetSugared(),
	}
}

func (h *KVHandler) Create(c *gin.Context) {
	var req request.CreateKVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errs.FormatValidationError(err),
		})
		return
	}

	h.log.Debugw("Creating key-value pair", "key", req.Key, "ttl", req.TTL)

	entry, err := h.service.Set(c.Request.Context(), req.Key, req.Value, req.TTL)
	if err != nil {
		h.log.Errorw("Failed to create key-value pair", "key", req.Key, "error", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.NewKVResponse(
		"Key-value pair stored successfully",
		entry.Key,
		"",
		entry.CreatedAt,
		entry.ExpiresAt,
	)

	h.log.Infow("Successfully created key-value pair", "key", entry.Key, "expires_at", entry.ExpiresAt)
	c.JSON(http.StatusCreated, resp)
}

func (h *KVHandler) Get(c *gin.Context) {
	key := c.Param("key")
	h.log.Debugw("Getting value", "key", key)

	entry, err := h.service.Get(c.Request.Context(), key)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrKeyNotFound) {
			status = http.StatusNotFound
			h.log.Warnw("Key not found", "key", key)
		} else {
			h.log.Errorw("Failed to get value", "key", key, "error", err)
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.NewKVResponse(
		"Key found",
		entry.Key,
		entry.Value,
		entry.CreatedAt,
		entry.ExpiresAt,
	)

	h.log.Infow("Successfully retrieved value", "key", key, "expires_at", entry.ExpiresAt)
	c.JSON(http.StatusOK, resp)
}

func (h *KVHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	h.log.Debugw("Attempting to delete key", "key", key)

	if err := h.service.Delete(c.Request.Context(), key); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrKeyNotFound) {
			status = http.StatusNotFound
			h.log.Warnw("Key not found for deletion", "key", key)
		} else {
			h.log.Errorw("Failed to delete key", "key", key, "error", err)
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	h.log.Infow("Successfully deleted key", "key", key)
	c.JSON(http.StatusOK, response.KVResponse{
		Message: "Key deleted successfully",
		Key:     key,
	})
}
