package handler

import (
	"errors"
	"key-value-store/internal/service"
	"key-value-store/internal/store"
	"key-value-store/internal/transport/http/handler/request"
	"key-value-store/internal/transport/http/handler/response"
	"key-value-store/internal/transport/http/middleware"
	"key-value-store/internal/util/http_util"
	"log/slog"
	"net/http"
)

type BucketHandler struct {
	service service.IBucketService
}

func NewBucketHandler(service service.IBucketService) *BucketHandler {
	return &BucketHandler{
		service: service,
	}
}

func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())

	var req request.CreateBucketRequest
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

	bucket, err := h.service.CreateBucket(r.Context(), req.Name, req.Description, req.ShardCount)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrBucketAlreadyExists):
			http_util.WriteConflict(w, "Bucket already exists")
		case errors.Is(err, store.ErrInvalidBucketName):
			http_util.WriteBadRequest(w, "Invalid bucket name")
		default:
			slog.Error("Handler: Failed to create bucket", "crr-id", crrid, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	resp := response.NewBucketResponseFromBucket(*bucket)
	http_util.WriteCreated(w, resp)
}

func (h *BucketHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	bucket, err := h.service.GetBucket(r.Context(), bucketName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrBucketNotFound):
			http_util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to get bucket", "crr-id", crrid, "bucket", bucketName, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	resp := response.NewBucketResponseFromBucket(*bucket)
	http_util.WriteOK(w, resp)
}

func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		http_util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	err := h.service.DeleteBucket(r.Context(), bucketName)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrBucketNotFound):
			http_util.WriteNotFound(w, "Bucket not found")
		default:
			slog.Error("Handler: Failed to delete bucket", "crr-id", crrid, "bucket", bucketName, "error", err)
			http_util.WriteInternalError(w)
		}
		return
	}

	http_util.WriteNoContent(w, "Bucket deleted successfully")
}

func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	crrid := middleware.CorrelationID(r.Context())

	buckets, err := h.service.ListBuckets(r.Context())
	if err != nil {
		slog.Error("Handler: Failed to list buckets", "crr-id", crrid, "error", err)
		http_util.WriteInternalError(w)
		return
	}

	resp := response.NewBucketListResponse(buckets)
	http_util.WriteOK(w, resp)
}
