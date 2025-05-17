package service

import (
	"context"
	"key-value-store/internal/config"
	"key-value-store/internal/errs"
	"key-value-store/internal/model"
	"key-value-store/internal/store"
	"key-value-store/internal/transport/http/middleware"
	"key-value-store/internal/util/compression"
	"log/slog"
	"time"
)

const (
	GCInterval = 5
)

type IStorageService interface {
	Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error)
	Get(ctx context.Context, key string) (model.StorageEntry, error)
	Delete(ctx context.Context, key string)
}

type storageService struct {
	store store.Store
	cfg   *config.Configs
}

func NewStorageService(store store.Store) IStorageService {
	s := &storageService{
		store: store,
		cfg:   config.Get(),
	}
	s.store.StartGC(GCInterval * time.Second)
	return s
}

func (s *storageService) Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to set key-value pair", "crr-id", crrid, "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl < 0 {
		slog.Debug("Service: Invalid TTL provided", "crr-id", crrid, "key", key, "ttl", ttl)
		return model.StorageEntry{}, errs.ErrInvalidTTL
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	entry := model.StorageEntry{
		Key:          key,
		Value:        value,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    exp,
		SingleRead:   singleRead,
		OriginalSize: int64(len(value)),
	}

	// Compress if value size is above threshold and compression is enabled
	if s.cfg.Store.CompressionType != "none" && int64(len(value)) > s.cfg.Store.CompressionThreshold {
		compressedValue, err := compression.CompressAndEncode([]byte(value), compression.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to compress value", "crr-id", crrid, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		entry.Value = compressedValue
		entry.Compressed = true
		entry.CompressedSize = int64(len(compressedValue))
		slog.Debug("Service: Value compressed", "crr-id", crrid, "key", key, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	}

	s.store.Set(key, entry)

	slog.Info("Service: Saved key-value pair", "crr-id", crrid, "key", key, "ttl", entry.TTL, "single_read", singleRead,
		"compressed", entry.Compressed, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, key string) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to get value", "crr-id", crrid, "key", key)

	entry, exists := s.store.Get(key)
	if !exists {
		slog.Debug("Service: Key not found in store", "crr-id", crrid, "key", key)
		return model.StorageEntry{}, errs.ErrKeyNotFound
	}

	if entry.IsExpired() {
		slog.Debug("Service: Key has expired in store", "crr-id", crrid, "key", key, "expires_at", entry.ExpiresAt.Format(time.RFC3339))
		s.store.Delete(key)
		return model.StorageEntry{}, errs.ErrKeyExpired
	}

	// Decompress if needed
	if entry.Compressed {
		decompressed, err := compression.DecodeAndDecompress(entry.Value, compression.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to decompress value", "crr-id", crrid, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		entry.Value = string(decompressed)
		entry.Compressed = false
		entry.CompressedSize = 0
		slog.Debug("Service: Value decompressed", "crr-id", crrid, "key", key)
	}

	if entry.SingleRead {
		s.store.Delete(key)
		slog.Debug("Service: Deleted single-read key after reading", "crr-id", crrid, "key", key)
	}

	slog.Debug("Service: Retrieved key-value pair", "crr-id", crrid, "key", key)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, key string) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to delete key", "crr-id", crrid, "key", key)
	s.store.Delete(key)
	slog.Info("Service: Deleted key", "crr-id", crrid, "key", key)
}
