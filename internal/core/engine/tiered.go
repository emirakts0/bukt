package engine

import (
	"key-value-store/internal/core/model"
	"key-value-store/internal/errs"
	"math"
	"math/rand"
	"sync"
	"time"

	"key-value-store/internal/config"

	"golang.org/x/sync/singleflight"
)

type TieredStore struct {
	memoryStore  Store
	diskStore    Store
	mu           sync.RWMutex
	evictionStop chan struct{}
	sf           singleflight.Group
	*GarbageCollector
}

func NewTieredStore(memoryStore, diskStore Store, cfg config.EngineConfig) Store {
	s := &TieredStore{
		memoryStore:  memoryStore,
		diskStore:    diskStore,
		evictionStop: make(chan struct{}),
	}
	s.GarbageCollector = NewGarbageCollector(s.Delete)
	s.StartEviction(cfg.EvictionInterval, cfg.EvictionBatchSize, 16) // 16 fixed sample size
	return s
}

func (s *TieredStore) Set(key string, entry model.StorageEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry.OnDisk = false
	s.memoryStore.Set(key, entry)
	s.GarbageCollector.Track(key, &entry)
}

func (s *TieredStore) Get(key string) (model.StorageEntry, bool) {
	entry, exists := s.memoryStore.Get(key)
	if !exists {
		return model.StorageEntry{}, false
	}

	if !entry.OnDisk {
		return entry, true
	}

	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		diskEntry, diskExists := s.diskStore.Get(key)
		if !diskExists {
			s.memoryStore.Delete(key)
			return nil, errs.ErrInconsistentState
		}

		diskEntry.OnDisk = false
		s.memoryStore.Set(key, diskEntry)

		go s.diskStore.Delete(key)
		return diskEntry, nil
	})

	if err != nil {
		return model.StorageEntry{}, false
	}

	return v.(model.StorageEntry), true
}

func (s *TieredStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.memoryStore.Delete(key)
	go s.diskStore.Delete(key)
	s.GarbageCollector.Untrack(key)
}

func (s *TieredStore) Exists(key string) bool {
	return s.memoryStore.Exists(key)
}

func (s *TieredStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keySet := make(map[string]struct{})

	memKeys := s.memoryStore.Keys()
	for _, key := range memKeys {
		keySet[key] = struct{}{}
	}

	diskKeys := s.diskStore.Keys()
	for _, key := range diskKeys {
		keySet[key] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	return keys
}

func (s *TieredStore) StartEviction(interval time.Duration, batchSize int, sampleSize int) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.evictLRU(batchSize, sampleSize)
			case <-s.evictionStop:
				return
			}
		}
	}()
}

func (s *TieredStore) StopEviction() {
	close(s.evictionStop)
}

func (s *TieredStore) evictLRU(batchSize int, sampleSize int) {
	keys := s.memoryStore.Keys()
	if len(keys) == 0 {
		return
	}

	toEvict := make(map[string]model.StorageEntry)

	for i := 0; i < batchSize && len(toEvict) < batchSize; i++ {
		var oldestKey string
		var oldestEntry model.StorageEntry
		oldestAccessTime := int64(math.MaxInt64)

		for j := 0; j < sampleSize; j++ {
			randIndex := rand.Intn(len(keys))
			randomKey := keys[randIndex]

			if _, alreadyFound := toEvict[randomKey]; alreadyFound {
				continue
			}

			entry, exists := s.memoryStore.Get(randomKey)
			if !exists || entry.OnDisk || entry.Value == nil {
				continue
			}

			if entry.LastAccess < oldestAccessTime {
				oldestAccessTime = entry.LastAccess
				oldestKey = randomKey
				oldestEntry = entry
			}
		}

		if oldestKey != "" {
			toEvict[oldestKey] = oldestEntry
		}
	}

	if len(toEvict) == 0 {
		return
	}

	for key, entry := range toEvict {
		s.diskStore.Set(key, entry)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for key, entry := range toEvict {
		currentEntry, exists := s.memoryStore.Get(key)
		if !exists || currentEntry.OnDisk || currentEntry.LastAccess != entry.LastAccess {
			go s.diskStore.Delete(key)
			continue
		}

		entry.Value = nil
		entry.OnDisk = true
		s.memoryStore.Set(key, entry)
	}
}

func (s *TieredStore) StartGC(interval time.Duration) {
	s.GarbageCollector.Start(interval)
}

func (s *TieredStore) StopGC() {
	s.StopEviction()
	s.GarbageCollector.Stop()
}

func (s *TieredStore) Usage() int64 {
	// TieredStore only reports the memory usage of its hot layer.
	return s.memoryStore.Usage()
}
