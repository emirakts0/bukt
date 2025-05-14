package response

import (
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
