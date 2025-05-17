package handler

import (
	"errors"
	"key-value-store/internal/errs"
	"key-value-store/internal/service"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"key-value-store/internal/transport/http/middleware"
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
	crrid := middleware.CorrelationID(r.Context())
	var req request.CreateKVRequest

	if err := http_util.ReadJSONBody(r, &req, w); err != nil {
		slog.Debug("Handler: Invalid JSON body", "crr-id", crrid, "error", err)
		http_util.WriteBadRequest(w, "Invalid JSON.")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Debug("Handler: Invalid request payload", "crr-id", crrid, "error", err)
		http_util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.service.Set(r.Context(), req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidTTL):
			http_util.WriteBadRequest(w, "Invalid TTL")
		default:
			http_util.WriteInternalError(w)
		}
		return
	}

	http_util.WriteCreated(w, "Key-value pair stored successfully")
}

func (h *KVHandler) Get(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())

	key := r.PathValue("key")
	if key == "" {
		slog.Debug("Handler: Key is required for Get request", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.service.Get(r.Context(), key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrKeyNotFound):
			http_util.WriteNotFound(w, "key not found")
		case errors.Is(err, errs.ErrKeyExpired):
			http_util.WriteBadRequest(w, "key is expired")
		default:
			http_util.WriteInternalError(w)
		}
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
	crrid := middleware.CorrelationID(r.Context())

	key := r.PathValue("key")
	if key == "" {
		slog.Debug("Handler: Key is required for Delete request", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	h.service.Delete(r.Context(), key)
	http_util.WriteNoContent(w, "No Content")
}
