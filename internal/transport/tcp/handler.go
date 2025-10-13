package tcp

import (
	"context"
	"errors"
	"key-value-store/internal/auth"
	"key-value-store/internal/errs"
	"key-value-store/internal/service"
	"log/slog"
)

type Handler struct {
	storageService service.IStorageService
	ctx            context.Context
}

func NewHandler(storageService service.IStorageService) *Handler {
	return &Handler{
		storageService: storageService,
		ctx:            context.Background(),
	}
}

func (h *Handler) HandleFrame(frame *Frame) *Frame {
	ctx := h.ctx

	switch frame.Command {
	case CmdSet:
		return h.handleSet(ctx, frame)
	case CmdGet:
		return h.handleGet(ctx, frame)
	case CmdDelete:
		return h.handleDelete(ctx, frame)
	default:
		return NewErrorFrame(frame.RequestID, StatusBadRequest, "Unknown command")
	}
}

func (h *Handler) handleSet(ctx context.Context, frame *Frame) *Frame {
	token, bucket, key, ttl, singleRead, value, err := DecodeSetPayload(frame.Payload)
	if err != nil {
		slog.Debug("TCP: Failed to decode SET payload", "error", err)
		return NewErrorFrame(frame.RequestID, StatusBadRequest, "Invalid payload")
	}

	if !auth.Manager().ValidateToken(token, bucket) {
		slog.Debug("TCP: Invalid token for SET", "bucket", bucket)
		return NewErrorFrame(frame.RequestID, StatusUnauthorized, "Invalid token")
	}

	_, err = h.storageService.Set(ctx, bucket, key, value, ttl, singleRead)
	if err != nil {
		return h.handleServiceError(frame.RequestID, err)
	}

	return NewResponseFrame(frame.RequestID, StatusCreated, nil)
}

func (h *Handler) handleGet(ctx context.Context, frame *Frame) *Frame {
	token, bucket, key, err := DecodeGetPayload(frame.Payload)
	if err != nil {
		slog.Debug("TCP: Failed to decode GET payload", "error", err)
		return NewErrorFrame(frame.RequestID, StatusBadRequest, "Invalid payload")
	}

	if !auth.Manager().ValidateToken(token, bucket) {
		slog.Debug("TCP: Invalid token for GET", "bucket", bucket)
		return NewErrorFrame(frame.RequestID, StatusUnauthorized, "Invalid token")
	}

	entry, err := h.storageService.Get(ctx, bucket, key)
	if err != nil {
		return h.handleServiceError(frame.RequestID, err)
	}

	responseData := EncodeValueResponse(
		entry.Key,
		entry.TTL,
		entry.CreatedAt.Unix(),
		entry.ExpiresAt.Unix(),
		entry.SingleRead,
		entry.Value,
	)

	return NewResponseFrame(frame.RequestID, StatusOK, responseData)
}

func (h *Handler) handleDelete(ctx context.Context, frame *Frame) *Frame {
	token, bucket, key, err := DecodeDeletePayload(frame.Payload)
	if err != nil {
		slog.Debug("TCP: Failed to decode DELETE payload", "error", err)
		return NewErrorFrame(frame.RequestID, StatusBadRequest, "Invalid payload")
	}

	if !auth.Manager().ValidateToken(token, bucket) {
		slog.Debug("TCP: Invalid token for DELETE", "bucket", bucket)
		return NewErrorFrame(frame.RequestID, StatusUnauthorized, "Invalid token")
	}

	err = h.storageService.Delete(ctx, bucket, key)
	if err != nil {
		return h.handleServiceError(frame.RequestID, err)
	}

	return NewResponseFrame(frame.RequestID, StatusNoContent, nil)
}

func (h *Handler) handleServiceError(requestID uint64, err error) *Frame {
	var status byte
	var message string

	switch {
	case errors.Is(err, errs.ErrInvalidTTL):
		status = StatusInvalidTTL
		message = "Invalid TTL"
	case errors.Is(err, errs.ErrKeyNotFound):
		status = StatusNotFound
		message = "Key not found"
	case errors.Is(err, errs.ErrKeyExpired):
		status = StatusKeyExpired
		message = "Key expired"
	case errors.Is(err, errs.ErrUnauthorized):
		status = StatusUnauthorized
		message = "Unauthorized"
	default:
		status = StatusInternalError
		message = "Internal server error"
		slog.Error("TCP: Unhandled service error", "error", err)
	}

	return NewErrorFrame(requestID, status, message)
}
