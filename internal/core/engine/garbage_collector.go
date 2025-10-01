package engine

import (
	"container/heap"
	"key-value-store/internal/core"
	"sync"
	"time"
)

type GarbageCollector struct {
	mu         sync.Mutex
	ttlHeap    ttlHeap
	ttlIndex   map[string]int
	stopCh     chan struct{}
	deleteFunc func(key string)
}

func NewGarbageCollector(deleteFunc func(key string)) *GarbageCollector {
	gc := &GarbageCollector{
		ttlHeap:    make(ttlHeap, 0),
		ttlIndex:   make(map[string]int),
		stopCh:     make(chan struct{}),
		deleteFunc: deleteFunc,
	}
	heap.Init(&gc.ttlHeap)
	return gc
}

func (gc *GarbageCollector) Track(key string, entry *core.StorageEntry) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if entry.TTL <= 0 {
		gc.Untrack(key)
		return
	}

	expireAt := entry.CreatedAt.Add(time.Duration(entry.TTL)).UnixNano()
	if itemIndex, exists := gc.ttlIndex[key]; exists {
		item := gc.ttlHeap[itemIndex]
		item.ExpireAt = expireAt
		heap.Fix(&gc.ttlHeap, item.Index)
	} else {
		item := &TtlHeapItem{
			Key:      key,
			ExpireAt: expireAt,
		}
		heap.Push(&gc.ttlHeap, item)
		gc.ttlIndex[key] = item.Index
	}
}

func (gc *GarbageCollector) Untrack(key string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if idx, ok := gc.ttlIndex[key]; ok {
		heap.Remove(&gc.ttlHeap, idx)
		delete(gc.ttlIndex, key)
	}
}

func (gc *GarbageCollector) collect() {
	var keysToDelete []string

	gc.mu.Lock()
	now := time.Now().UnixNano()
	for gc.ttlHeap.Len() > 0 {
		item := gc.ttlHeap[0]
		if item.ExpireAt > now {
			break
		}

		poppedItem := heap.Pop(&gc.ttlHeap).(*TtlHeapItem)
		delete(gc.ttlIndex, poppedItem.Key)
		keysToDelete = append(keysToDelete, poppedItem.Key)
	}
	gc.mu.Unlock()

	for _, key := range keysToDelete {
		if gc.deleteFunc != nil {
			gc.deleteFunc(key)
		}
	}
}

func (gc *GarbageCollector) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				gc.collect()
			case <-gc.stopCh:
				return
			}
		}
	}()
}

func (gc *GarbageCollector) Stop() {
	close(gc.stopCh)
}
