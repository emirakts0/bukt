package engine

import "time"

type Store interface {
	Set(key string, entry StorageEntry)
	Get(key string) (StorageEntry, bool)
	Delete(key string)
	Exists(key string) bool
	Keys() []string
	StartGC(interval time.Duration)
	StopGC()
	Usage() int64
	Count() int64
}
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
	AccessCount    int32
	LastAccess     int64
}

func (e StorageEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}
