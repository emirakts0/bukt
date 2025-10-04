package response

import (
	"key-value-store/internal/bucket"
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
	AuthToken   string `json:"auth_token,omitempty"`
}

type BucketListResponse struct {
	Buckets []BucketResponse `json:"buckets"`
	Count   int              `json:"count"`
}

func NewBucketResponseFromMetadata(meta *bucket.BucketMetadata, token string) BucketResponse {
	return BucketResponse{
		ID:          meta.ID,
		Name:        meta.Name,
		Description: meta.Description,
		CreatedAt:   meta.CreatedAt.Format(time.RFC3339),
		ShardCount:  meta.ShardCount,
		KeyCount:    meta.KeyCount,
		MemoryUsage: meta.MemoryUsage,
		AuthToken:   token,
	}
}

func NewBucketListResponse(buckets []*bucket.BucketMetadata) BucketListResponse {
	bucketResponses := make([]BucketResponse, len(buckets))
	for i, b := range buckets {
		bucketResponses[i] = NewBucketResponseFromMetadata(b, "")
	}

	return BucketListResponse{
		Buckets: bucketResponses,
		Count:   len(buckets),
	}
}
