package response

import (
	"key-value-store/internal/core/model"
	"time"
)

type KVResponse struct {
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewKVResponse(key, value string, createdAt, expiresAt time.Time) KVResponse {
	return KVResponse{
		Key:       key,
		Value:     value,
		CreatedAt: createdAt.Format(time.RFC3339),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}
}

// NewKVResponseFromEntry creates a KVResponse from a StorageEntry
func NewKVResponseFromEntry(entry model.StorageEntry) KVResponse {
	// Convert byte array to string for JSON response
	// Simple string conversion - API remains backward compatible
	valueStr := string(entry.Value)

	return KVResponse{
		Key:       entry.Key,
		Value:     valueStr,
		CreatedAt: entry.CreatedAt.Format(time.RFC3339),
		ExpiresAt: entry.ExpiresAt.Format(time.RFC3339),
	}
}
