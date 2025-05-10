package store

import (
	"key-value-store/internal/model"
	"sync"
)

// MemoryStore is responsible only for storing and retrieving data
type MemoryStore interface {
	Set(key string, entry model.StorageEntry)
	Get(key string) (model.StorageEntry, bool)
	Delete(key string) bool
	Exists(key string) bool
}

type memoryStore struct {
	store map[string]model.StorageEntry
	mu    sync.RWMutex
}

// NewMemoryStore creates a new instance of memory store
func NewMemoryStore() MemoryStore {
	return &memoryStore{
		store: make(map[string]model.StorageEntry),
	}
}

func (s *memoryStore) Set(key string, entry model.StorageEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = entry
}

func (s *memoryStore) Get(key string) (model.StorageEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, exists := s.store[key]
	return entry, exists
}

func (s *memoryStore) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.store[key]; !exists {
		return false
	}
	delete(s.store, key)
	return true
}

func (s *memoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.store[key]
	return exists
}
