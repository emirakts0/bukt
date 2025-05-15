package request

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxKeyLength = 255
	MinKeyLength = 1
	MinTTL       = 1
)

type CreateKVRequest struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	TTL        int64  `json:"ttl"` // in seconds
	SingleRead bool   `json:"single_read"`
}

func (r *CreateKVRequest) Validate() error {
	r.Sanitize()

	if r.Key == "" {
		return errors.New("key is required")
	}
	if len(r.Key) < MinKeyLength {
		return fmt.Errorf("key must be at least %d character", MinKeyLength)
	}
	if len(r.Key) > MaxKeyLength {
		return fmt.Errorf("key must be at most %d characters", MaxKeyLength)
	}
	if r.Value == "" {
		return errors.New("value is required")
	}
	if r.TTL < MinTTL {
		return fmt.Errorf("ttl must be greater than %d", MinTTL-1)
	}

	return nil
}

// Sanitize performs any necessary sanitization on the request
func (r *CreateKVRequest) Sanitize() {
	r.Key = strings.TrimSpace(r.Key)
	r.Value = strings.TrimSpace(r.Value)
}
