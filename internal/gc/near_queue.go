package gc

import (
	"container/heap"
	"key-value-store/internal/util"
	"sync"
)

type nearEntry struct {
	key      string
	expireAt int64
	index    int
}

type nearHeap []*nearEntry

func (h nearHeap) Len() int           { return len(h) }
func (h nearHeap) Less(i, j int) bool { return h[i].expireAt < h[j].expireAt }
func (h nearHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *nearHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*nearEntry)
	item.index = n
	*h = append(*h, item)
}

func (h *nearHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

type NearQueue struct {
	mu       sync.Mutex
	heap     nearHeap
	index    map[string]int
	windowMs int64
}

func NewNearQueue(windowMs int64) *NearQueue {
	nq := &NearQueue{
		heap:     make(nearHeap, 0),
		index:    make(map[string]int),
		windowMs: windowMs,
	}
	heap.Init(&nq.heap)
	return nq
}

func (nq *NearQueue) Add(key string, expireAt int64) bool {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	now := util.CachedNow() / 1e6
	if expireAt/1e6-now > nq.windowMs {
		return false
	}

	if idx, exists := nq.index[key]; exists {
		e := nq.heap[idx]
		e.expireAt = expireAt
		heap.Fix(&nq.heap, e.index)
	} else {
		e := &nearEntry{key: key, expireAt: expireAt}
		heap.Push(&nq.heap, e)
		nq.index[key] = e.index
	}
	return true
}

func (nq *NearQueue) Remove(key string) {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	if idx, ok := nq.index[key]; ok {
		heap.Remove(&nq.heap, idx)
		delete(nq.index, key)
	}
}

func (nq *NearQueue) DrainExpired() []string {
	nq.mu.Lock()
	defer nq.mu.Unlock()

	now := util.CachedNow() / 1e6
	var expired []string

	for nq.heap.Len() > 0 {
		e := nq.heap[0]
		if e.expireAt/1e6 > now {
			break
		}
		popped := heap.Pop(&nq.heap).(*nearEntry)
		delete(nq.index, popped.key)
		expired = append(expired, popped.key)
	}

	return expired
}
