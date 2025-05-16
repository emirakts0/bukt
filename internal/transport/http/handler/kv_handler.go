package handler

import (
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"key-value-store/internal/util/http_util"
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

	if err := http_util.ReadJSONBody(r, &req, w); err != nil {
		http_util.WriteBadRequest(w, "Invalid request.")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Warn("Invalid request", "error", err)
		http_util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.service.Set(r.Context(), req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		http_util.WriteBadRequest(w, err.Error())
		return
	}

	http_util.WriteCreated(w, "Key-value pair stored successfully")
}

func (h *KVHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.service.Get(r.Context(), key)
	if err != nil {
		//todo handlerda handle ederken türe daha cok dikkat edilebilir.
		http_util.WriteNotFound(w, err.Error())
		return
	}

	http_util.WriteOK(w, response.NewKVResponse(
		entry.Key,
		entry.Value,
		entry.CreatedAt,
		entry.ExpiresAt,
	))
}

func (h *KVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	if err := h.service.Delete(r.Context(), key); err != nil {
		http_util.WriteNotFound(w, err.Error())
		return
	}

	http_util.WriteOK(w, "Key deleted successfully")
}
