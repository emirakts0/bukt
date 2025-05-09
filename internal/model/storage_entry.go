package model

import "time"

type StorageEntry struct {
	Key       string
	Value     string
	ExpiresAt time.Time
}

func (e StorageEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}
