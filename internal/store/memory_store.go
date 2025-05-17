package store

import (
	"key-value-store/internal/model"
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

type memoryStore struct {
	store       map[string]model.StorageEntry
	mu          sync.RWMutex // for store operations
	usedMemory  int64
	accessCount map[string]*int64
	lastAccess  map[string]*int64
	gcStop      chan struct{}
}

func NewMemoryStore() Store {
	return &memoryStore{
		store:       make(map[string]model.StorageEntry),
		accessCount: make(map[string]*int64),
		lastAccess:  make(map[string]*int64),
		gcStop:      make(chan struct{}),
	}
}

func (s *memoryStore) StartGC(interval time.Duration) {
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

func (s *memoryStore) StopGC() {
	close(s.gcStop)
}

func (s *memoryStore) cleanupExpiredKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.store {
		if entry.IsExpired() {
			s.usedMemory -= int64(len(key) + len(entry.Value))
			delete(s.store, key)
			delete(s.accessCount, key)
			delete(s.lastAccess, key)
		}
	}
}

func (s *memoryStore) Set(key string, entry model.StorageEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entrySize := int64(len(key) + len(entry.Value))

	// Check old
	if oldEntry, exists := s.store[key]; exists {
		s.usedMemory -= int64(len(key) + len(oldEntry.Value))
	}
	s.usedMemory += entrySize

	s.accessCount[key] = new(int64)
	s.lastAccess[key] = new(int64)
	*s.lastAccess[key] = time.Now().UnixNano()

	s.store[key] = entry
}

func (s *memoryStore) Get(key string) (model.StorageEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]

	if exists {
		if _, ok := s.accessCount[key]; !ok {
			s.accessCount[key] = new(int64)
			s.lastAccess[key] = new(int64)
		}
		atomic.AddInt64(s.accessCount[key], 1)
		atomic.StoreInt64(s.lastAccess[key], time.Now().UnixNano())
	}
	return entry, exists
}

func (s *memoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[key]
	if !exists {
		return
	}

	s.usedMemory -= int64(len(key) + len(entry.Value))

	delete(s.store, key)
	delete(s.accessCount, key)
	delete(s.lastAccess, key)
	return
}

func (s *memoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.store[key]
	return exists
}

func (s *memoryStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.store))
	for key := range s.store {
		keys = append(keys, key)
	}
	return keys
}

func (s *memoryStore) GetMemoryUsage() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.usedMemory
}
