package service

import (
	"context"
	"key-value-store/internal/config"
	"key-value-store/internal/errs"
	"key-value-store/internal/model"
	"key-value-store/internal/store"
	"key-value-store/internal/transport/http/middleware"
	"log/slog"
	"time"
)

const (
	GCInterval = 5
)

type IStorageService interface {
	Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error)
	Get(ctx context.Context, key string) (model.StorageEntry, error)
	Delete(ctx context.Context, key string) error
}

type storageService struct {
	store store.IShardedStore
	cfg   *config.Config
}

func NewStorageService(store store.IShardedStore) IStorageService {
	s := &storageService{
		store: store,
		cfg:   config.Get(),
	}
	s.store.StartGC(GCInterval * time.Second)
	return s
}

func (s *storageService) Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)

	slog.Debug("Attempting to set key-value pair", "crr-id", crrid, "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl <= 0 {
		slog.Warn("Invalid TTL provided", "crr-id", crrid, "key", key, "ttl", ttl)
		return model.StorageEntry{}, errs.ErrInvalidTTL
	}

	if s.store.Exists(key) {
		slog.Warn("Key already exists", "crr-id", crrid, "key", key)
		return model.StorageEntry{}, errs.ErrKeyAlreadyExists
	}

	now := time.Now()
	entry := model.StorageEntry{
		Key:          key,
		Value:        value,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Duration(ttl) * time.Second),
		SingleRead:   singleRead,
		OriginalSize: int64(len(value)),
	}

	// Compress if value size is above threshold and compression is enabled
	if s.cfg.Store.CompressionType != "none" &&
		int64(len(value)) > s.cfg.Store.CompressionThreshold {
		if err := entry.CompressValue(s.cfg.Store.CompressionType); err != nil {
			slog.Error("Failed to compress value", "crr-id", crrid, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		slog.Debug("Value compressed", "crr-id", crrid, "key", key, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	}

	if err := s.store.Set(key, entry); err != nil {
		slog.Error("Failed to set key-value pair", "crr-id", crrid, "key", key, "error", err)
		return model.StorageEntry{}, errs.ErrMemoryLimit
	}

	slog.Info("Successfully set key-value pair", "crr-id", crrid, "key", key, "ttl", entry.TTL, "single_read", singleRead,
		"compressed", entry.Compressed, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, key string) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)

	slog.Debug("Attempting to get value", "crr-id", crrid, "key", key)

	entry, exists := s.store.Get(key)
	if !exists {
		slog.Warn("Key not found", "crr-id", crrid, "key", key)
		return model.StorageEntry{}, errs.ErrKeyNotFound
	}

	if entry.IsExpired() {
		slog.Warn("Key has expired", "crr-id", crrid, "key", key, "expires_at", entry.ExpiresAt.Format(time.RFC3339))
		err := s.store.Delete(key)
		if err != nil {
			slog.Error("Deletion error", "crr-id", crrid, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrDeletion
		}
		return model.StorageEntry{}, errs.ErrKeyExpired
	}

	// Decompress if needed
	if entry.Compressed {
		if err := entry.DecompressValue(config.Get().Store.CompressionType); err != nil {
			slog.Error("Failed to decompress value", "crr-id", crrid, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		slog.Debug("Value decompressed", "crr-id", crrid, "key", key)
	}

	if entry.SingleRead {
		err := s.store.Delete(key)
		if err != nil {
			return model.StorageEntry{}, errs.ErrDeletion
		}
		slog.Info("Deleted single-read key after reading", "crr-id", crrid, "key", key, "single_read", entry.SingleRead)
	}

	slog.Info("Successfully retrieved value", "crr-id", crrid, "key", key, "single_read", entry.SingleRead,
		"compressed", entry.Compressed, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, key string) error {
	crrid := middleware.CorrelationID(ctx)

	slog.Debug("Attempting to delete key", "crr-id", crrid, "key", key)

	if err := s.store.Delete(key); err != nil {
		slog.Warn("Key not found", "crr-id", crrid, "key", key, "err", err)
		return err
	}

	slog.Info("Successfully deleted key", "crr-id", crrid, "key", key)
	return nil
}
