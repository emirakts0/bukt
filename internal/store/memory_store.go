package store

import (
	"key-value-store/internal/errs"
	"key-value-store/internal/model"
	"sync"
	"time"
)

type MemoryStore interface {
	Set(key string, entry model.StorageEntry) error
	Get(key string) (model.StorageEntry, bool)
	Delete(key string) error
	Exists(key string) bool
	Keys() []string
	StartGC(interval time.Duration)
	StopGC()
	GetMemoryUsage() int64
}

type memoryStore struct {
	store       map[string]model.StorageEntry
	keyLocks    sync.Map     // map[string]*sync.Mutex
	mu          sync.RWMutex // for store operations
	maxMemory   int64
	usedMemory  int64
	accessCount map[string]int64
	lastAccess  map[string]time.Time
	gcStop      chan struct{}
}

func NewMemoryStore() MemoryStore {
	return &memoryStore{
		store:       make(map[string]model.StorageEntry),
		maxMemory:   100 * 1024 * 1024, // 100MB default limit
		accessCount: make(map[string]int64),
		lastAccess:  make(map[string]time.Time),
		gcStop:      make(chan struct{}),
	}
}

// getKeyLock returns a mutex for the given key, creating it if it doesn't exist
func (s *memoryStore) getKeyLock(key string) *sync.Mutex {
	value, _ := s.keyLocks.LoadOrStore(key, &sync.Mutex{})
	return value.(*sync.Mutex)
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
			keyLock := s.getKeyLock(key)
			keyLock.Lock()

			s.usedMemory -= int64(len(key) + len(entry.Value))
			delete(s.store, key)
			delete(s.accessCount, key)
			delete(s.lastAccess, key)
			s.keyLocks.Delete(key)

			keyLock.Unlock()
		}
	}
}

func (s *memoryStore) Set(key string, entry model.StorageEntry) error {
	keyLock := s.getKeyLock(key)
	keyLock.Lock()
	defer keyLock.Unlock()

	s.mu.Lock()
	// Calculate entry size
	entrySize := int64(len(key) + len(entry.Value))

	// Check old
	if oldEntry, exists := s.store[key]; exists {
		s.usedMemory -= int64(len(key) + len(oldEntry.Value))
	}
	s.usedMemory += entrySize
	s.mu.Unlock()

	s.accessCount[key] = 0
	s.lastAccess[key] = time.Now()

	s.store[key] = entry
	return nil
}

func (s *memoryStore) Get(key string) (model.StorageEntry, bool) {
	keyLock := s.getKeyLock(key)
	keyLock.Lock()
	defer keyLock.Unlock()

	entry, exists := s.store[key]
	if exists {
		s.accessCount[key]++
		s.lastAccess[key] = time.Now()
	}
	return entry, exists
}

func (s *memoryStore) Delete(key string) error {
	keyLock := s.getKeyLock(key)
	keyLock.Lock()
	defer keyLock.Unlock()

	entry, exists := s.store[key]
	if !exists {
		return errs.ErrKeyNotFound
	}

	s.mu.Lock()
	s.usedMemory -= int64(len(key) + len(entry.Value))
	s.mu.Unlock()

	delete(s.store, key)
	delete(s.accessCount, key)
	delete(s.lastAccess, key)
	s.keyLocks.Delete(key)
	return nil
}

func (s *memoryStore) Exists(key string) bool {
	keyLock := s.getKeyLock(key)
	keyLock.Lock()
	defer keyLock.Unlock()

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
