package engine

import (
	"key-value-store/internal/core/model"
	"sync"
	"sync/atomic"
	"time"
)

type Store interface {
	Set(key string, entry model.StorageEntry)
	Get(key string) (model.StorageEntry, bool)
	Delete(key string)
	Exists(key string) bool
	Keys() []string
	StartGC(interval time.Duration)
	StopGC()
	GetMemoryUsage() int64
}

type MemoryStore struct {
	store      map[string]*model.StorageEntry
	mu         sync.RWMutex // for store operations
	usedMemory int64
	gcStop     chan struct{}
}

func NewMemoryStore() Store {
	return &MemoryStore{
		store:  make(map[string]*model.StorageEntry),
		gcStop: make(chan struct{}),
	}
}

func (s *MemoryStore) StartGC(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpiredKeys()
			case <-s.gcStop:
				return
			}
		}
	}()
}

func (s *MemoryStore) StopGC() {
	close(s.gcStop)
}

func (s *MemoryStore) cleanupExpiredKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.store {
		if entry.IsExpired() {
			s.usedMemory -= int64(len(key) + len(entry.Value))
			delete(s.store, key)
		}
	}
}

func (s *MemoryStore) Set(key string, entry model.StorageEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entrySize := int64(len(key) + len(entry.Value))

	// Check old
	if oldEntry, exists := s.store[key]; exists {
		s.usedMemory -= int64(len(key) + len(oldEntry.Value))
	}
	s.usedMemory += entrySize

	entry.LastAccess = time.Now().UnixNano()
	entry.AccessCount = 0

	s.store[key] = &entry
}

func (s *MemoryStore) Get(key string) (model.StorageEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]
	if !exists {
		return model.StorageEntry{}, false
	}

	if entry.IsExpired() {
		go s.Delete(key)
		return model.StorageEntry{}, false
	}

	if entry.SingleRead {
		if atomic.AddInt64(&entry.AccessCount, 1) > 1 {
			return model.StorageEntry{}, false
		}

		go s.Delete(key)
		return *entry, true
	}

	atomic.AddInt64(&entry.AccessCount, 1)
	atomic.StoreInt64(&entry.LastAccess, time.Now().UnixNano())
	return *entry, exists
}

func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[key]
	if !exists {
		return
	}

	s.usedMemory -= int64(len(key) + len(entry.Value))

	delete(s.store, key)
	return
}

func (s *MemoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.store[key]
	return exists
}

func (s *MemoryStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.store))
	for key := range s.store {
		keys = append(keys, key)
	}
	return keys
}

func (s *MemoryStore) GetMemoryUsage() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.usedMemory
}
