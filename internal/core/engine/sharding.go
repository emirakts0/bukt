package engine

import (
	"fmt"
	"hash/fnv"
	"key-value-store/internal/config"
	"key-value-store/internal/core/model"
	"path/filepath"
	"sync"
	"time"
)

type ShardContainer struct {
	shards      []Store
	shardCount  int
	useSharding bool
}

func NewShardContainer(bucketName string, shardCount int, cfg config.EngineConfig) *ShardContainer {
	storeFactory := func(shardIndex int) Store {
		diskFilePath := filepath.Join(cfg.DataDir, fmt.Sprintf("%s_shard_%d.db", bucketName, shardIndex))

		switch cfg.Type {
		case "tiered":
			memoryStore := NewMemoryStore()
			diskStore, err := NewDiskStore(diskFilePath)
			if err != nil {
				panic(fmt.Sprintf("failed to create disk store for bucket %s, shard %d: %v", bucketName, shardIndex, err))
			}
			tieredStore := NewTieredStore(memoryStore, diskStore, cfg)
			return tieredStore
		default:
			return NewMemoryStore()
		}
	}

	if shardCount <= 1 {
		return &ShardContainer{
			shards:      []Store{storeFactory(0)},
			shardCount:  1,
			useSharding: false,
		}
	}

	shards := make([]Store, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = storeFactory(i)
	}

	return &ShardContainer{
		shards:      shards,
		shardCount:  shardCount,
		useSharding: true,
	}
}

func (sc *ShardContainer) getShard(key string) Store {
	if !sc.useSharding {
		return sc.shards[0]
	}

	hash := fnv.New32a()
	_, err := hash.Write([]byte(key))
	if err != nil {
		panic(err)
	}
	return sc.shards[hash.Sum32()%uint32(sc.shardCount)]
}

func (sc *ShardContainer) Set(key string, entry model.StorageEntry) {
	shard := sc.getShard(key)
	shard.Set(key, entry)
}

func (sc *ShardContainer) Get(key string) (model.StorageEntry, bool) {
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
	if sc.shardCount == 1 {
		return sc.shards[0].Keys()
	}

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

func (sc *ShardContainer) Usage() int64 {
	var totalUsage int64

	if sc.shardCount == 1 {
		return sc.shards[0].Usage()
	}

	for _, shard := range sc.shards {
		totalUsage += shard.Usage()
	}
	return totalUsage
}
