package response

import (
	"key-value-store/internal/util"
	"time"
)

type KVResponse struct {
	Message   string `json:"message"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewKVResponse(message, key, value string, createdAt, expiresAt time.Time) KVResponse {
	formatter := util.NewTimeFormatter()
	return KVResponse{
		Message:   message,
		Key:       key,
		Value:     value,
		CreatedAt: formatter.FormatTime(createdAt),
		ExpiresAt: formatter.FormatTime(expiresAt),
	}
}
