package service

import (
	"context"
	"encoding/hex"
	"key-value-store/internal/bucket"
	"key-value-store/internal/config"
	"key-value-store/internal/engine"
	"key-value-store/internal/errs"
	"key-value-store/internal/transport/http/middleware"
	"key-value-store/internal/util"
	"log/slog"
	"time"
)

type IStorageService interface {
	Set(ctx context.Context, bucketName, authTokenHex, key, value string, ttl int64, singleRead bool) (engine.StorageEntry, error)
	Get(ctx context.Context, bucketName, authTokenHex, key string) (engine.StorageEntry, error)
	Delete(ctx context.Context, bucketName, authTokenHex, key string) error
}

type storageService struct {
	bucketManager bucket.BucketManager
	cfg           *config.Configuration
}

func NewStorageService(bucketManager bucket.BucketManager, cfg *config.Configuration) IStorageService {
	s := &storageService{
		bucketManager: bucketManager,
		cfg:           cfg,
	}
	return s
}

func (s *storageService) Set(ctx context.Context, bucketName, authTokenHex, key, value string, ttl int64, singleRead bool) (engine.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to set key-value pair in bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl < 0 {
		slog.Debug("Service: Invalid TTL provided", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl)
		return engine.StorageEntry{}, errs.ErrInvalidTTL
	}

	tokenBytes, err := hex.DecodeString(authTokenHex)
	if err != nil || len(tokenBytes) != 16 {
		slog.Error("Service: Invalid token format", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrUnauthorized
	}

	bucketStore, ok := s.bucketManager.AuthenticateAndGetStore(bucketName, tokenBytes)
	if !ok {
		slog.Error("Service: Failed to authenticate bucket", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrUnauthorized
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	valueBytes := []byte(value)
	originalSize := int64(len(valueBytes))

	entry := engine.StorageEntry{
		Key:          key,
		Value:        valueBytes,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    exp,
		SingleRead:   singleRead,
		OriginalSize: originalSize,
	}

	if s.cfg.Store.CompressionType != "none" && originalSize > s.cfg.Store.CompressionThreshold {
		compressedValue, err := util.CompressBytes(valueBytes, util.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to compress value", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			return engine.StorageEntry{}, errs.ErrCompression
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

func (s *storageService) Get(ctx context.Context, bucketName, authTokenHex, key string) (engine.StorageEntry, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to get value from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	tokenBytes, err := hex.DecodeString(authTokenHex)
	if err != nil || len(tokenBytes) != 16 {
		slog.Error("Service: Invalid token format", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrUnauthorized
	}

	bucketStore, ok := s.bucketManager.AuthenticateAndGetStore(bucketName, tokenBytes)
	if !ok {
		slog.Error("Service: Failed to authenticate bucket", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrUnauthorized
	}

	entry, exists := bucketStore.Get(key)
	if !exists {
		slog.Debug("Service: Key not found in bucket engine", "crr-id", crrid, "bucket", bucketName, "key", key)
		return engine.StorageEntry{}, errs.ErrKeyNotFound
	}

	if entry.Compressed {
		decompressed, err := util.DecompressBytes(entry.Value, util.CompressionType(s.cfg.Store.CompressionType))
		if err != nil {
			slog.Error("Service: Failed to decompress value", "crr-id", crrid, "bucket", bucketName, "key", key, "error", err)
			return engine.StorageEntry{}, errs.ErrCompression
		}
		entry.Value = decompressed
		entry.Compressed = false
		entry.CompressedSize = 0
		slog.Debug("Service: Value decompressed", "crr-id", crrid, "bucket", bucketName, "key", key)
	}

	if entry.SingleRead {
		slog.Debug("Service: Deleted single-read key after reading", "crr-id", crrid, "bucket", bucketName, "key", key)
	}

	slog.Debug("Service: Retrieved key-value pair from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, bucketName, authTokenHex, key string) error {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("Service: Attempting to delete key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	tokenBytes, err := hex.DecodeString(authTokenHex)
	if err != nil || len(tokenBytes) != 16 {
		slog.Error("Service: Invalid token format", "crr-id", crrid, "bucket", bucketName)
		return errs.ErrUnauthorized
	}

	bucketStore, ok := s.bucketManager.AuthenticateAndGetStore(bucketName, tokenBytes)
	if !ok {
		slog.Error("Service: Failed to authenticate bucket", "crr-id", crrid, "bucket", bucketName)
		return errs.ErrUnauthorized
	}

	bucketStore.Delete(key)
	slog.Info("Service: Deleted key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return nil
}
