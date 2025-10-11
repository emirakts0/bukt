package handler

import (
	"errors"
	"key-value-store/internal/errs"
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
	crrid := util.GetCorrelationID(r.Context())

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	var req request.CreateKVRequest
	if err := util.ReadJSONBody(r, &req, w); err != nil {
		slog.Debug("Handler: Invalid JSON body", "crr-id", crrid, "error", err)
		util.WriteBadRequest(w, "Invalid JSON.")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Debug("Handler: Invalid request payload", "crr-id", crrid, "error", err)
		util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.service.Set(r.Context(), bucketName, req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidTTL):
			util.WriteBadRequest(w, "Invalid TTL")
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid bucket auth token")
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to set key-value in bucket", "crr-id", crrid, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	util.WriteCreated(w, "Key-value pair stored successfully in bucket")
}

func (h *KVHandler) Get(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	key := r.PathValue("key")

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required for Get request", "crr-id", crrid)
		util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.service.Get(r.Context(), bucketName, key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrKeyNotFound):
			util.WriteNotFound(w, "Key not found")
		case errors.Is(err, errs.ErrKeyExpired):
			util.WriteNotFound(w, "Key expired")
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid bucket auth token")
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to get key from bucket", "crr-id", crrid, "key", key, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	resp := response.NewKVResponseFromEntry(entry)
	util.WriteOK(w, resp)
}

func (h *KVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	key := r.PathValue("key")

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required for Delete request", "crr-id", crrid)
		util.WriteBadRequest(w, "Key is required")
		return
	}

	err := h.service.Delete(r.Context(), bucketName, key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid bucket auth token")
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to delete key from bucket", "crr-id", crrid, "key", key, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	util.WriteNoContent(w, "Key deleted successfully from bucket")
}
