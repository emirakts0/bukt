package service

import (
	"context"
	"key-value-store/internal/core"
	"key-value-store/internal/transport/http/middleware"
	"log/slog"
)

type IBucketService interface {
	CreateBucket(ctx context.Context, name, description string, shardCount int) (*core.Bucket, error)
	GetBucket(ctx context.Context, name string) (*core.Bucket, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]core.Bucket, error)
}

type bucketService struct {
	bucketManager core.BucketManager
}

func NewBucketService(bucketManager core.BucketManager) IBucketService {
	return &bucketService{
		bucketManager: bucketManager,
	}
}

func (s *bucketService) CreateBucket(ctx context.Context, name, description string, shardCount int) (*core.Bucket, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("BucketService: Creating bucket", "crr-id", crrid, "name", name, "shard_count", shardCount)

	bucket, err := s.bucketManager.CreateBucket(name, description, shardCount)
	if err != nil {
		slog.Error("BucketService: Failed to create bucket", "crr-id", crrid, "name", name, "error", err)
		return nil, err
	}

	slog.Info("BucketService: Created bucket", "crr-id", crrid, "name", name, "shard_count", shardCount)
	return bucket, nil
}

func (s *bucketService) GetBucket(ctx context.Context, name string) (*core.Bucket, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("BucketService: Getting bucket", "crr-id", crrid, "name", name)

	bucket, err := s.bucketManager.GetBucket(name)
	if err != nil {
		slog.Debug("BucketService: Bucket not found", "crr-id", crrid, "name", name)
		return nil, err
	}

	return bucket, nil
}

func (s *bucketService) DeleteBucket(ctx context.Context, name string) error {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("BucketService: Deleting bucket", "crr-id", crrid, "name", name)

	err := s.bucketManager.DeleteBucket(name)
	if err != nil {
		slog.Error("BucketService: Failed to delete bucket", "crr-id", crrid, "name", name, "error", err)
		return err
	}

	slog.Info("BucketService: Deleted bucket", "crr-id", crrid, "name", name)
	return nil
}

func (s *bucketService) ListBuckets(ctx context.Context) ([]core.Bucket, error) {
	crrid := middleware.CorrelationID(ctx)
	slog.Debug("BucketService: Listing buckets", "crr-id", crrid)

	buckets := s.bucketManager.ListBuckets()
	return buckets, nil
}
