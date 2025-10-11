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

type BucketHandler struct {
	service service.IBucketService
}

func NewBucketHandler(service service.IBucketService) *BucketHandler {
	return &BucketHandler{
		service: service,
	}
}

func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())

	var req request.CreateBucketRequest
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

	result, err := h.service.CreateBucket(r.Context(), req.Name, req.Description, req.ShardCount)
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

	resp := response.NewBucketResponseFromMetadata(result.Metadata, result.AuthToken)
	util.WriteCreated(w, resp)
}

func (h *BucketHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	meta, err := h.service.GetBucket(r.Context(), bucketName)
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

	resp := response.NewBucketResponseFromMetadata(meta, "")
	util.WriteOK(w, resp)
}

func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())
	bucketName := r.PathValue("bucket")

	if bucketName == "" {
		slog.Debug("Handler: Bucket name is required", "crr-id", crrid)
		util.WriteBadRequest(w, "Bucket name is required")
		return
	}

	var req request.DeleteBucketRequest
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

	err := h.service.DeleteBucket(r.Context(), bucketName, req.AuthToken)
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

func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	crrid := util.GetCorrelationID(r.Context())

	buckets, err := h.service.ListBuckets(r.Context())
	if err != nil {
		slog.Error("Handler: Failed to list buckets", "crr-id", crrid, "error", err)
		util.WriteInternalError(w)
		return
	}

	resp := response.NewBucketListResponse(buckets)
	util.WriteOK(w, resp)
}
