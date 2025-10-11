package service

import (
	"context"
	"key-value-store/internal/bucket"
	"key-value-store/internal/config"
	"key-value-store/internal/engine"
	"key-value-store/internal/errs"
	"key-value-store/internal/util"
	"log/slog"
	"time"
)

type IStorageService interface {
	Set(ctx context.Context, bucketName, key string, value []byte, ttl int64, singleRead bool) (engine.StorageEntry, error)
	Get(ctx context.Context, bucketName, key string) (engine.StorageEntry, error)
	Delete(ctx context.Context, bucketName, key string) error
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

func (s *storageService) Set(ctx context.Context, bucketName, key string, value []byte, ttl int64, singleRead bool) (engine.StorageEntry, error) {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("Service: Attempting to set key-value pair in bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl < 0 {
		slog.Debug("Service: Invalid TTL provided", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", ttl)
		return engine.StorageEntry{}, errs.ErrInvalidTTL
	}

	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrBucketNotFound
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	originalSize := int64(len(value))

	entry := engine.StorageEntry{
		Key:          key,
		Value:        value,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    exp,
		SingleRead:   singleRead,
		OriginalSize: originalSize,
	}

	bucketStore.Set(key, entry)

	slog.Info("Service: Saved key-value pair in bucket", "crr-id", crrid, "bucket", bucketName, "key", key, "ttl", entry.TTL, "single_read", singleRead, "original_size", entry.OriginalSize)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, bucketName, key string) (engine.StorageEntry, error) {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("Service: Attempting to get value from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrBucketNotFound
	}

	entry, exists := bucketStore.Get(key)
	if !exists {
		slog.Debug("Service: Key not found in bucket engine", "crr-id", crrid, "bucket", bucketName, "key", key)
		return engine.StorageEntry{}, errs.ErrKeyNotFound
	}

	if entry.SingleRead {
		slog.Debug("Service: Deleted single-read key after reading", "crr-id", crrid, "bucket", bucketName, "key", key)
	}

	slog.Debug("Service: Retrieved key-value pair from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, bucketName, key string) error {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("Service: Attempting to delete key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)

	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return errs.ErrBucketNotFound
	}

	bucketStore.Delete(key)
	slog.Info("Service: Deleted key from bucket", "crr-id", crrid, "bucket", bucketName, "key", key)
	return nil
}
