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
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	var req request.CreateKVRequest
	if err := http_util.ReadJSONBody(r, &req, w); err != nil {
		slog.Debug("Handler: Invalid JSON body", "crr-id", crrid, "bucket", bucketName, "error", err)
		http_util.WriteBadRequest(w, "Invalid JSON.")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Debug("Handler: Invalid request payload", "crr-id", crrid, "bucket", bucketName, "error", err)
		http_util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.service.Set(r.Context(), bucketName, req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidTTL):
			http_util.WriteBadRequest(w, "Invalid TTL")
		case errors.Is(err, errs.ErrBucketNotFound):
			http_util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to set key-value in bucket", "crr-id", crrid, "bucket", bucketName, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	http_util.WriteCreated(w, "Key-value pair stored successfully in bucket")
}

func (h *KVHandler) Get(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())
	bucketName := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required for Get request", "crr-id", crrid, "bucket", bucketName)
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.service.Get(r.Context(), bucketName, key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrKeyNotFound):
			http_util.WriteNotFound(w, "Key not found")
		case errors.Is(err, errs.ErrKeyExpired):
			http_util.WriteNotFound(w, "Key expired")
		case errors.Is(err, errs.ErrBucketNotFound):
			http_util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to get key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	resp := response.NewKVResponseFromEntry(entry)
	http_util.WriteOK(w, resp)
}

func (h *KVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())
	bucketName := r.PathValue("bucket")
	key := r.PathValue("key")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required for Delete request", "crr-id", crrid, "bucket", bucketName)
		http_util.WriteBadRequest(w, "Key is required")
		return
	}

	err := h.service.Delete(r.Context(), bucketName, key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrBucketNotFound):
			http_util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to delete key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	http_util.WriteNoContent(w, "Key deleted successfully from bucket")
}
