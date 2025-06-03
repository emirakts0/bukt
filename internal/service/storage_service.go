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

type IStorageService interface {
	Set(ctx context.Context, bucketName, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error)
	Get(ctx context.Context, bucketName, key string) (model.StorageEntry, error)
	Delete(ctx context.Context, bucketName, key string) error
}

type storageService struct {
	bucketManager store.BucketManager
	cfg           *config.Configuration
}

func NewStorageService(bucketManager store.BucketManager) IStorageService {
	s := &storageService{
		bucketManager: bucketManager,
		cfg:           config.Config(),
	}
	return s
}

func (s *storageService) Set(ctx context.Context, bucketName, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to set key-value pair in bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl < 0 {
		slog.Debug("Service: Invalid TTL provided", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl)
		return model.StorageEntry{}, errs.ErrInvalidTTL
	}

	// Get bucket store from bucket manager
	bucketStore, err := s.bucketManager.GetBucketStore(bucketName)
	if err != nil {
		slog.Error("Service: Failed to get bucket store", "crr-id", crrid, "bucket", bucketName, "error", err)
		return model.StorageEntry{}, err
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	// Convert string to bytes immediately
	valueBytes := []byte(value)
	originalSize := int64(len(valueBytes))

	entry := model.StorageEntry{
		Key:          key,
		Value:        valueBytes,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    exp,
		SingleRead:   singleRead,
		OriginalSize: originalSize,
	}

	// Compress if value size is above threshold and compression is enabled
	if s.cfg.Store.CompressionType != "none" && originalSize > s.cfg.Store.CompressionThreshold {
		compressedValue, err := compression.CompressBytes(valueBytes, compression.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to compress value", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		entry.Value = compressedValue
		entry.Compressed = true
		entry.CompressedSize = int64(len(compressedValue))
		slog.Debug("Service: Value compressed", "crr-id", crrid, "bucket", bucketName, "key", key, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	}

	bucketStore.Set(key, entry)

	slog.Info("Service: Saved key-value pair in bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", entry.TTL, "single_read", singleRead,
		"compressed", entry.Compressed, "original_size", entry.OriginalSize, "compressed_size", entry.CompressedSize)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, bucketName, key string) (model.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to get value from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	// Get bucket store from bucket manager
	bucketStore, err := s.bucketManager.GetBucketStore(bucketName)
	if err != nil {
		slog.Error("Service: Failed to get bucket store", "crr-id", crrid, "bucket", bucketName, "error", err)
		return model.StorageEntry{}, err
	}

	entry, exists := bucketStore.Get(key)
	if !exists {
		slog.Debug("Service: Key not found in bucket store", "crr-id", crrid, "bucket", bucketName, "key", key)
		return model.StorageEntry{}, errs.ErrKeyNotFound
	}

	if entry.IsExpired() {
		slog.Debug("Service: Key has expired in bucket store", "crr-id", crrid, "bucket", bucketName, "key", key, "expires_at", entry.ExpiresAt.Format(time.RFC3339))
		bucketStore.Delete(key)
		return model.StorageEntry{}, errs.ErrKeyExpired
	}

	// Decompress if needed
	if entry.Compressed {
		decompressed, err := compression.DecompressBytes(entry.Value, compression.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to decompress value", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			return model.StorageEntry{}, errs.ErrCompression
		}
		entry.Value = decompressed
		entry.Compressed = false
		entry.CompressedSize = 0
		slog.Debug("Service: Value decompressed", "crr-id", crrid, "bucket", bucketName, "key", key)
	}

	if entry.SingleRead {
		bucketStore.Delete(key)
		slog.Debug("Service: Deleted single-read key after reading", "crr-id", crrid, "bucket", bucketName, "key", key)
	}

	slog.Debug("Service: Retrieved key-value pair from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, bucketName, key string) error {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to delete key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	bucketStore, err := s.bucketManager.GetBucketStore(bucketName)
	if err != nil {
		slog.Error("Service: Failed to get bucket store", "crr-id", crrid, "bucket", bucketName, "error", err)
		return err
	}

	bucketStore.Delete(key)
	slog.Info("Service: Deleted key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return nil
}
