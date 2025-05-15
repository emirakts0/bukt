package handler

import (
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"key-value-store/internal/util"
	"log/slog"
	"net/http"
)

type KVHandler struct {
	service service.IStorageService
}

func NewKVHandler(service service.IStorageService) *KVHandler {
	return &KVHandler{
		service: service,
	}
}

func (h *KVHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateKVRequest

	if err := util.ReadJSONBody(r, &req); err != nil {
		util.WriteBadRequest(w, "Failed to read request body")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Warn("Invalid request", "error", err)
		util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.service.Set(r.Context(), req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		util.WriteBadRequest(w, err.Error())
		return
	}

	util.WriteCreated(w, "Key-value pair stored successfully")
}

func (h *KVHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.service.Get(r.Context(), key)
	if err != nil {
		util.WriteNotFound(w, err.Error())
		return
	}

	util.WriteOK(w, response.NewKVResponse(
		entry.Key,
		entry.Value,
		entry.CreatedAt,
		entry.ExpiresAt,
	))
}

func (h *KVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		util.WriteBadRequest(w, "Key is required")
		return
	}

	if err := h.service.Delete(r.Context(), key); err != nil {
		util.WriteNotFound(w, err.Error())
		return
	}

	util.WriteOK(w, map[string]string{
		"message": "Key deleted successfully",
	})
}
