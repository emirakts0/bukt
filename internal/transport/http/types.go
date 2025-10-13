package http

import (
	"errors"
	"key-value-store/internal/bucket"
	"key-value-store/internal/engine"
	"key-value-store/internal/util"
	"strings"
	"time"
)

// Request types
type CreateKVRequest struct {
	Key        string `json:"key"`
	Value      []byte `json:"value"`
	TTL        int64  `json:"ttl"`
	SingleRead bool   `json:"single_read"`
}

type CreateBucketRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ShardCount  int    `json:"shard_count,omitempty"`
}

type DeleteBucketRequest struct {
	AuthToken string `json:"auth_token"`
}

// Response types
type KVResponse struct {
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

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

// Validation methods
func (r *CreateKVRequest) Validate() error {
	r.Key = strings.TrimSpace(r.Key)
	if r.Key == "" {
		return errors.New("key is required")
	}
	if len(r.Key) > 255 {
		return errors.New("key too long (max 255)")
	}
	if len(r.Value) == 0 {
		return errors.New("value is required")
	}
	if r.TTL < 0 {
		return errors.New("ttl must be non-negative")
	}
	return nil
}

func (r *CreateBucketRequest) Validate() error {
	r.Name = strings.TrimSpace(strings.ToLower(r.Name))
	r.Description = strings.TrimSpace(r.Description)

	if r.Name == "" {
		return errors.New("bucket name is required")
	}
	if len(r.Name) > 63 {
		return errors.New("bucket name too long (max 63)")
	}
	if len(r.Description) > 256 {
		return errors.New("description too long (max 256)")
	}
	if r.ShardCount < 0 {
		return errors.New("shard count must be non-negative")
	}

	// Simple validation: lowercase, numbers, hyphens
	for _, c := range r.Name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return errors.New("bucket name must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

func (r *DeleteBucketRequest) Validate() error {
	if r.AuthToken == "" {
		return errors.New("auth token is required")
	}
	return nil
}

// Response helpers
func kvResponseFromEntry(entry engine.StorageEntry) KVResponse {
	return KVResponse{
		Key:       entry.Key,
		Value:     util.BytesToString(entry.Value),
		CreatedAt: entry.CreatedAt.Format(time.RFC3339),
		ExpiresAt: entry.ExpiresAt.Format(time.RFC3339),
	}
}

func bucketResponse(meta *bucket.BucketMetadata, token string) BucketResponse {
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

func bucketListResponse(buckets []*bucket.BucketMetadata) BucketListResponse {
	responses := make([]BucketResponse, len(buckets))
	for i, b := range buckets {
		responses[i] = bucketResponse(b, "")
	}
	return BucketListResponse{
		Buckets: responses,
		Count:   len(buckets),
	}
}
