package store

import (
	"hash/fnv"
	"key-value-store/internal/model"
	"sync"
	"time"
)

type IShardedStore interface {
	Set(key string, entry model.StorageEntry) error
	Get(key string) (model.StorageEntry, bool)
	Delete(key string) error
	Exists(key string) bool
	Keys() []string
	StartGC(interval time.Duration)
	StopGC()
}

type shardedStore struct {
	shards      []MemoryStore
	shardCount  int
	useSharding bool
}

func NewShardedStore(shardCount int) IShardedStore {
	if shardCount <= 1 {
		// Single shard case
		return &shardedStore{
			shards:      []MemoryStore{NewMemoryStore()},
			shardCount:  1,
			useSharding: false,
		}
	}

	shards := make([]MemoryStore, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = NewMemoryStore()
	}

	return &shardedStore{
		shards:      shards,
		shardCount:  shardCount,
		useSharding: true,
	}
}

func (s *shardedStore) getShard(key string) MemoryStore {
	if !s.useSharding {
		return s.shards[0]
	}

	hash := fnv.New32a()
	_, err := hash.Write([]byte(key))
	if err != nil {
		panic(err)
	}
	return s.shards[hash.Sum32()%uint32(s.shardCount)]
}

func (s *shardedStore) Set(key string, entry model.StorageEntry) error {
	shard := s.getShard(key)
	return shard.Set(key, entry)
}

func (s *shardedStore) Get(key string) (model.StorageEntry, bool) {
	shard := s.getShard(key)
	return shard.Get(key)
}

func (s *shardedStore) Delete(key string) error {
	shard := s.getShard(key)
	return shard.Delete(key)
}

func (s *shardedStore) Exists(key string) bool {
	shard := s.getShard(key)
	return shard.Exists(key)
}

func (s *shardedStore) Keys() []string {
	resultChan := make(chan []string, len(s.shards))
	var wg sync.WaitGroup

	for _, shard := range s.shards {
		wg.Add(1)
		go func(s MemoryStore) {
			defer wg.Done()
			resultChan <- s.Keys()
		}(shard)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allKeys []string
	for keys := range resultChan {
		allKeys = append(allKeys, keys...)
	}

	return allKeys
}

func (s *shardedStore) StartGC(interval time.Duration) {
	for _, shard := range s.shards {
		shard.StartGC(interval)
	}
}

func (s *shardedStore) StopGC() {
	for _, shard := range s.shards {
		shard.StopGC()
	}
}
