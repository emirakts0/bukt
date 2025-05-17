package store

import (
	"hash/fnv"
	"key-value-store/internal/model"
	"sync"
	"time"
)

// Coordinator manages multiple Store instances (shards).
type Coordinator struct {
	shards      []Store
	shardCount  int
	useSharding bool
}

func NewCoordinator(shardCount int) Store {
	if shardCount <= 1 {
		return &Coordinator{
			shards:      []Store{NewMemoryStore()},
			shardCount:  1,
			useSharding: false,
		}
	}

	shards := make([]Store, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = NewMemoryStore()
	}

	return &Coordinator{
		shards:      shards,
		shardCount:  shardCount,
		useSharding: true,
	}
}

func (c *Coordinator) getShard(key string) Store {
	if !c.useSharding {
		return c.shards[0]
	}

	hash := fnv.New32a()
	_, err := hash.Write([]byte(key))
	if err != nil {
		// handle
		panic(err)
	}
	return c.shards[hash.Sum32()%uint32(c.shardCount)]
}

func (c *Coordinator) Set(key string, entry model.StorageEntry) {
	shard := c.getShard(key)
	shard.Set(key, entry)
}

func (c *Coordinator) Get(key string) (model.StorageEntry, bool) {
	shard := c.getShard(key)
	return shard.Get(key)
}

func (c *Coordinator) Delete(key string) {
	shard := c.getShard(key)
	shard.Delete(key)
}

func (c *Coordinator) Exists(key string) bool {
	shard := c.getShard(key)
	return shard.Exists(key)
}

func (c *Coordinator) Keys() []string {
	if c.shardCount == 1 {
		return c.shards[0].Keys()
	}

	resultChan := make(chan []string, len(c.shards))
	var wg sync.WaitGroup

	for _, shard := range c.shards {
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

func (c *Coordinator) StartGC(interval time.Duration) {
	for _, shard := range c.shards {
		shard.StartGC(interval)
	}
}

func (c *Coordinator) StopGC() {
	for _, shard := range c.shards {
		shard.StopGC()
	}
}

func (c *Coordinator) GetMemoryUsage() int64 {
	var totalMemoryUsage int64

	if c.shardCount == 1 {
		return c.shards[0].GetMemoryUsage()
	}

	for _, shard := range c.shards {
		totalMemoryUsage += shard.GetMemoryUsage()
	}
	return totalMemoryUsage
}
