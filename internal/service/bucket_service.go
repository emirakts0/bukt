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
	tokenHex, err := s.bucketManager.CreateBucket(name, description, shardCount)
	if err != nil {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("BucketService: Failed to create bucket", "crr-id", crrid, "name", name, "error", err)
		return nil, err
	}

	meta, ok := s.bucketManager.GetBucket(name)
	if !ok {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("BucketService: Bucket not found after creation", "crr-id", crrid, "name", name)
		return nil, errs.ErrBucketNotFound
	}

	return &CreateBucketResult{
		Metadata:  meta,
		AuthToken: tokenHex,
	}, nil
}

func (s *bucketService) GetBucket(ctx context.Context, name string) (*bucket.BucketMetadata, error) {
	meta, ok := s.bucketManager.GetBucket(name)
	if !ok {
		return nil, errs.ErrBucketNotFound
	}

	return meta, nil
}

func (s *bucketService) DeleteBucket(ctx context.Context, name, token string) error {
	err := s.bucketManager.DeleteBucket(name, token)
	if err != nil {
		crrid := util.GetCorrelationID(ctx)
		slog.Error("BucketService: Failed to delete bucket", "crr-id", crrid, "name", name, "error", err)
		return err
	}

	return nil
}

func (s *bucketService) ListBuckets(ctx context.Context) ([]*bucket.BucketMetadata, error) {
	buckets := s.bucketManager.ListBuckets()
	return buckets, nil
}
