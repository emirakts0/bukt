package model

import (
	"encoding/base64"
	"key-value-store/internal/util/compression"
	"time"
)

type StorageEntry struct {
	Key            string
	Value          string
	Compressed     bool
	CompressedSize int64
	OriginalSize   int64
	TTL            int64
	CreatedAt      time.Time
	ExpiresAt      time.Time
	SingleRead     bool
}

func (e StorageEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// CompressValue compresses the value if it's not already compressed
func (e *StorageEntry) CompressValue(compressionType string) error {
	if compressionType == "none" {
		return nil
	}

	if e.Compressed {
		return nil
	}

	// Compress the value
	compressed, err := compression.Compress([]byte(e.Value), compression.CompressionType(compressionType))
	if err != nil {
		return err
	}

	// Encode to base64
	e.Value = base64.StdEncoding.EncodeToString(compressed)
	e.Compressed = true
	e.CompressedSize = int64(len(e.Value))

	return nil
}

// DecompressValue decompresses the value if it's compressed
func (e *StorageEntry) DecompressValue(compressionType string) error {
	if compressionType == "none" || !e.Compressed {
		return nil
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(e.Value)
	if err != nil {
		return err
	}

	// Decompress
	decompressed, err := compression.Decompress(decoded, compression.CompressionType(compressionType))
	if err != nil {
		return err
	}

	e.Value = string(decompressed)
	e.Compressed = false
	e.CompressedSize = 0

	return nil
}
