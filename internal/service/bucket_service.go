package service

import (
	"context"
	"key-value-store/internal/bucket"
	"key-value-store/internal/errs"
	"key-value-store/internal/util"
	"log/slog"
)

type CreateBucketResult struct {
	Metadata  *bucket.BucketMetadata
	AuthToken string
}

type IBucketService interface {
	CreateBucket(ctx context.Context, name, description string, shardCount int) (*CreateBucketResult, error)
	GetBucket(ctx context.Context, name string) (*bucket.BucketMetadata, error)
	DeleteBucket(ctx context.Context, name, token string) error
	ListBuckets(ctx context.Context) ([]*bucket.BucketMetadata, error)
}

type bucketService struct {
	bucketManager bucket.BucketManager
}

func NewBucketService(bucketManager bucket.BucketManager) IBucketService {
	return &bucketService{
		bucketManager: bucketManager,
	}
}

func (s *bucketService) CreateBucket(ctx context.Context, name, description string, shardCount int) (*CreateBucketResult, error) {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("BucketService: Creating bucket", "crr-id", crrid, "name", name, "shard_count", shardCount)

	tokenHex, err := s.bucketManager.CreateBucket(name, description, shardCount)
	if err != nil {
		slog.Error("BucketService: Failed to create bucket", "crr-id", crrid, "name", name, "error", err)
		return nil, err
	}

	meta, ok := s.bucketManager.GetBucket(name)
	if !ok {
		slog.Error("BucketService: Bucket not found after creation", "crr-id", crrid, "name", name)
		return nil, errs.ErrBucketNotFound
	}

	slog.Info("BucketService: Created bucket", "crr-id", crrid, "name", name, "shard_count", shardCount)
	return &CreateBucketResult{
		Metadata:  meta,
		AuthToken: tokenHex,
	}, nil
}

func (s *bucketService) GetBucket(ctx context.Context, name string) (*bucket.BucketMetadata, error) {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("BucketService: Getting bucket", "crr-id", crrid, "name", name)

	meta, ok := s.bucketManager.GetBucket(name)
	if !ok {
		slog.Debug("BucketService: Bucket not found", "crr-id", crrid, "name", name)
		return nil, errs.ErrBucketNotFound
	}

	return meta, nil
}

func (s *bucketService) DeleteBucket(ctx context.Context, name, token string) error {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("BucketService: Deleting bucket", "crr-id", crrid, "name", name)

	err := s.bucketManager.DeleteBucket(name, token)
	if err != nil {
		slog.Error("BucketService: Failed to delete bucket", "crr-id", crrid, "name", name, "error", err)
		return err
	}

	slog.Info("BucketService: Deleted bucket", "crr-id", crrid, "name", name)
	return nil
}

func (s *bucketService) ListBuckets(ctx context.Context) ([]*bucket.BucketMetadata, error) {
	crrid := util.GetCorrelationID(ctx)
	slog.Debug("BucketService: Listing buckets", "crr-id", crrid)

	buckets := s.bucketManager.ListBuckets()
	return buckets, nil
}
