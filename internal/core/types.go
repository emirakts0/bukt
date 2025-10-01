package core

import "time"

type StorageEntry struct {
	Key            string
	Value          []byte
	Compressed     bool
	CompressedSize int64
	OriginalSize   int64
	TTL            int64
	CreatedAt      time.Time
	ExpiresAt      time.Time
	SingleRead     bool
	AccessCount    int64
	LastAccess     int64
	OnDisk         bool
}

func (e StorageEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}
