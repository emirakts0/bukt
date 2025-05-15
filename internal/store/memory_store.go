package store

import (
	"key-value-store/internal/model"
	"sync"
)

type MemoryStore interface {
	Set(key string, entry model.StorageEntry)
	Get(key string) (model.StorageEntry, bool)
	Delete(key string) bool
	Exists(key string) bool
	Keys() []string
}

type memoryStore struct {
	store map[string]model.StorageEntry
	mu    sync.RWMutex
}

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

func (s *memoryStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.store))
	for key := range s.store {
		keys = append(keys, key)
	}
	return keys
}
