package request

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxBucketNameLength  = 63
	MinBucketNameLength  = 1
	MaxDescriptionLength = 256
)

type CreateBucketRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ShardCount  int    `json:"shard_count,omitempty"`
}

func (r *CreateBucketRequest) Validate() error {
	r.Sanitize()

	if r.Name == "" {
		return errors.New("bucket name is required")
	}
	if len(r.Name) < MinBucketNameLength {
		return fmt.Errorf("bucket name must be at least %d character", MinBucketNameLength)
	}
	if len(r.Name) > MaxBucketNameLength {
		return fmt.Errorf("bucket name must be at most %d characters", MaxBucketNameLength)
	}
	if len(r.Description) > MaxDescriptionLength {
		return fmt.Errorf("description must be at most %d characters", MaxDescriptionLength)
	}
	if r.ShardCount < 0 {
		return errors.New("shard count must be non-negative")
	}

	// Validate bucket name format (similar to S3 bucket naming)
	if !isValidBucketName(r.Name) {
		return errors.New("invalid bucket name format")
	}

	return nil
}

// Sanitize performs any necessary sanitization on the request
func (r *CreateBucketRequest) Sanitize() {
	r.Name = strings.TrimSpace(strings.ToLower(r.Name))
	r.Description = strings.TrimSpace(r.Description)
}

// isValidBucketName validates bucket name format
func isValidBucketName(name string) bool {
	if len(name) < MinBucketNameLength || len(name) > MaxBucketNameLength {
		return false
	}

	// Must start and end with lowercase letter or number
	if !isAlphaNumeric(name[0]) || !isAlphaNumeric(name[len(name)-1]) {
		return false
	}

	// Can contain lowercase letters, numbers, and hyphens
	for _, char := range name {
		if !isAlphaNumeric(byte(char)) && char != '-' {
			return false
		}
	}

	// Cannot contain consecutive hyphens
	if strings.Contains(name, "--") {
		return false
	}

	return true
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
