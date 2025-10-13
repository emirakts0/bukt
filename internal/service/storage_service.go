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
	if ttl < 0 {
		return engine.StorageEntry{}, errs.ErrInvalidTTL
	}

	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrBucketNotFound
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(time.Duration(ttl) * time.Second)
	}

	entry := engine.StorageEntry{
		Key:          key,
		Value:        value,
		TTL:          ttl,
		CreatedAt:    now,
		ExpiresAt:    exp,
		SingleRead:   singleRead,
		OriginalSize: int64(len(value)),
	}

	bucketStore.Set(key, entry)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, bucketName, key string) (engine.StorageEntry, error) {
	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return engine.StorageEntry{}, errs.ErrBucketNotFound
	}

	entry, exists := bucketStore.Get(key)
	if !exists {
		return engine.StorageEntry{}, errs.ErrKeyNotFound
	}

	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, bucketName, key string) error {
	bucketStore, ok := s.bucketManager.GetStore(bucketName)
	if !ok {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("Service: Bucket not found", "crr-id", crrid, "bucket", bucketName)
		return errs.ErrBucketNotFound
	}

	bucketStore.Delete(key)
	return nil
}
