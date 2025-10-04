package engine

import (
	"key-value-store/internal/util"
	"sync"
	"time"
)

type ShardContainer struct {
	shards     []Store
	shardCount int
	hasher     util.Hasher
}

func NewShardContainer(shardCount int) *ShardContainer {
	if shardCount < 1 {
		shardCount = 1
	}

	shards := make([]Store, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = NewMemoryStore()
	}

	return &ShardContainer{
		shards:     shards,
		shardCount: shardCount,
		hasher:     util.NewDefaultHasher(),
	}
}

func (sc *ShardContainer) getShard(key string) Store {
	hash := sc.hasher.Sum64String(key)
	return sc.shards[hash%uint64(sc.shardCount)]
}

func (sc *ShardContainer) Set(key string, entry StorageEntry) {
	shard := sc.getShard(key)
	shard.Set(key, entry)
}

func (sc *ShardContainer) Get(key string) (StorageEntry, bool) {
	shard := sc.getShard(key)
	return shard.Get(key)
}

func (sc *ShardContainer) Delete(key string) {
	shard := sc.getShard(key)
	shard.Delete(key)
}

func (sc *ShardContainer) Exists(key string) bool {
	shard := sc.getShard(key)
	return shard.Exists(key)
}

func (sc *ShardContainer) Keys() []string {
	resultChan := make(chan []string, len(sc.shards))
	var wg sync.WaitGroup

	for _, shard := range sc.shards {
		wg.Add(1)
		go func(s Store) {
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

func (sc *ShardContainer) StartGC(interval time.Duration) {
	for _, shard := range sc.shards {
		shard.StartGC(interval)
	}
}

func (sc *ShardContainer) StopGC() {
	for _, shard := range sc.shards {
		shard.StopGC()
	}
}

func (sc *ShardContainer) Close() error {
	for _, shard := range sc.shards {
		shard.StopGC()
	}
	return nil
}

func (sc *ShardContainer) Count() int64 {
	var total int64
	for _, shard := range sc.shards {
		total += shard.Count()
	}
	return total
}

func (sc *ShardContainer) Usage() int64 {
	var totalUsage int64
	for _, shard := range sc.shards {
		totalUsage += shard.Usage()
	}
	return totalUsage
}
