package response

import (
	"key-value-store/internal/store"
	"time"
)

type BucketResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	ShardCount  int    `json:"shard_count"`
	KeyCount    int64  `json:"key_count"`
	MemoryUsage int64  `json:"memory_usage"`
}

type BucketListResponse struct {
	Buckets []BucketResponse `json:"buckets"`
	Count   int              `json:"count"`
}

func NewBucketResponseFromBucket(bucket store.Bucket) BucketResponse {
	return BucketResponse{
		ID:          bucket.ID,
		Name:        bucket.Name,
		Description: bucket.Description,
		CreatedAt:   bucket.CreatedAt.Format(time.RFC3339),
		ShardCount:  bucket.ShardCount,
		KeyCount:    bucket.KeyCount,
		MemoryUsage: bucket.MemoryUsage,
	}
}

func NewBucketListResponse(buckets []store.Bucket) BucketListResponse {
	bucketResponses := make([]BucketResponse, len(buckets))
	for i, bucket := range buckets {
		bucketResponses[i] = NewBucketResponseFromBucket(bucket)
	}

	return BucketListResponse{
		Buckets: bucketResponses,
		Count:   len(buckets),
	}
}
