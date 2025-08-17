package engine

import (
	"key-value-store/internal/core/model"
	"sync"
	"time"
)

// TODO: This implementation will be replaced with BadgerDB.
type DiskStore struct {
	mu sync.RWMutex
}

func NewDiskStore(filePath string) (Store, error) {
	return &DiskStore{}, nil
}

func (s *DiskStore) Set(key string, entry model.StorageEntry) {
}

func (s *DiskStore) Get(key string) (model.StorageEntry, bool) {
	return model.StorageEntry{}, false
}

func (s *DiskStore) Delete(key string) {
}

func (s *DiskStore) Exists(key string) bool {
	return false
}

func (s *DiskStore) Keys() []string {
	return nil
}

func (s *DiskStore) StartGC(interval time.Duration) {
}

func (s *DiskStore) StopGC() {
}

func (s *DiskStore) GetMemoryUsage() int64 {
	return 0
}
