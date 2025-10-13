package http

import (
	"errors"
	"key-value-store/internal/errs"
	"key-value-store/internal/service"
	"key-value-store/internal/util"
	"log/slog"
	"net/http"
)

type Handlers struct {
	storageService service.IStorageService
	bucketService  service.IBucketService
}

func NewHandlers(storageService service.IStorageService, bucketService service.IBucketService) *Handlers {
	return &Handlers{
		storageService: storageService,
		bucketService:  bucketService,
	}
}

// KV Handlers
func (h *Handlers) CreateKV(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	var req CreateKVRequest
	if err := util.ReadJSONBody(r, &req, w); err != nil {
		util.WriteBadRequest(w, "Invalid JSON")
		return
	}

	if err := req.Validate(); err != nil {
		util.WriteBadRequest(w, err.Error())
		return
	}

	_, err := h.storageService.Set(r.Context(), bucketName, req.Key, req.Value, req.TTL, req.SingleRead)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidTTL):
			util.WriteBadRequest(w, "Invalid TTL")
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid bucket auth token")
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to set key-value", "crr-id", crrid, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	util.WriteCreated(w, "Key-value pair stored successfully")
}

func (h *Handlers) GetKV(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	key := r.PathValue("key")

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Key is required")
		return
	}

	entry, err := h.storageService.Get(r.Context(), bucketName, key)
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
			slog.Error("Handler: Failed to get key", "crr-id", crrid, "key", key, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	resp := kvResponseFromEntry(entry)
	util.WriteOK(w, resp)
}

func (h *Handlers) DeleteKV(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	key := r.PathValue("key")

	bucketName, ok := util.GetBucketName(r.Context())
	if !ok {
		slog.Error("Handler: Bucket name not found in context", "crr-id", crrid)
		util.WriteUnauthorized(w, "Unauthorized")
		return
	}

	if key == "" {
		slog.Debug("Handler: Key is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Key is required")
		return
	}

	err := h.storageService.Delete(r.Context(), bucketName, key)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid bucket auth token")
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to delete key", "crr-id", crrid, "key", key, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	util.WriteNoContent(w, "Key deleted successfully")
}

// Bucket Handlers
func (h *Handlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())

	var req CreateBucketRequest
	if err := util.ReadJSONBody(r, &req, w); err != nil {
		util.WriteBadRequest(w, "Invalid JSON")
		return
	}

	if err := req.Validate(); err != nil {
		util.WriteBadRequest(w, err.Error())
		return
	}

	result, err := h.bucketService.CreateBucket(r.Context(), req.Name, req.Description, req.ShardCount)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrBucketAlreadyExists):
			util.WriteConflict(w, "Bucket already exists")
		case errors.Is(err, errs.ErrInvalidBucketName):
			util.WriteBadRequest(w, "Invalid bucket name")
		default:
			slog.Error("Handler: Failed to create bucket", "crr-id", crrid, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	resp := bucketResponse(result.Metadata, result.AuthToken)
	util.WriteCreated(w, resp)
}

func (h *Handlers) GetBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	meta, err := h.bucketService.GetBucket(r.Context(), bucketName)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to get bucket", "crr-id", crrid, "bucket", bucketName, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	resp := bucketResponse(meta, "")
	util.WriteOK(w, resp)
}

func (h *Handlers) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	var req DeleteBucketRequest
	if err := util.ReadJSONBody(r, &req, w); err != nil {
		slog.Debug("Handler: Invalid JSON body", "crr-id", crrid, "error", err)
		util.WriteBadRequest(w, "Invalid JSON")
		return
	}

	if err := req.Validate(); err != nil {
		slog.Debug("Handler: Invalid request", "crr-id", crrid, "error", err)
		util.WriteBadRequest(w, err.Error())
		return
	}

	err := h.bucketService.DeleteBucket(r.Context(), bucketName, req.AuthToken)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrBucketNotFound):
			util.WriteNotFound(w, "Bucket not found")
		case errors.Is(err, errs.ErrUnauthorized):
			util.WriteUnauthorized(w, "Invalid auth token")
		case errors.Is(err, errs.ErrCannotDeleteDefault):
			util.WriteBadRequest(w, "Cannot delete default bucket")
		default:
			slog.Error("Handler: Failed to delete bucket", "crr-id", crrid, "bucket", bucketName, "error", err)
			util.WriteInternalError(w)
		}
		return
	}

	util.WriteNoContent(w, "Bucket deleted successfully")
}

func (h *Handlers) ListBuckets(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())

	buckets, err := h.bucketService.ListBuckets(r.Context())
	if err != nil {
		slog.Error("Handler: Failed to list buckets", "crr-id", crrid, "error", err)
		util.WriteInternalError(w)
		return
	}

	resp := bucketListResponse(buckets)
	util.WriteOK(w, resp)
}
