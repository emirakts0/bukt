package handler

import (
	_ "context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"key-value-store/internal/errs"
	"key-value-store/internal/logger"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"key-value-store/internal/util"
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

	entry, err := h.service.Set(c.Request.Context(), req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := fmt.Sprintf("Key-value pair stored successfully. Key: %s, Value: %s, CreatedAt: %s, ExpiresAt: %s",
		entry.Key,
		entry.Value,
		util.NewTimeFormatter().FormatTime(entry.CreatedAt),
		util.NewTimeFormatter().FormatTime(entry.ExpiresAt),
	)

	c.JSON(http.StatusCreated, resp)
}

func (h *KVHandler) Get(c *gin.Context) {
	key := c.Param("key")

	entry, err := h.service.Get(c.Request.Context(), key)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrKeyNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	resp := response.NewKVResponse(
		entry.Key,
		entry.Value,
		entry.CreatedAt,
		entry.ExpiresAt,
	)

	c.JSON(http.StatusOK, resp)
}

func (h *KVHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	if err := h.service.Delete(c.Request.Context(), key); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrKeyNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Key deleted successfully Key: %s", key))
}
