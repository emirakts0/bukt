package gc

import (
	"sync"
	"time"
)

type DeleteCoalescer struct {
	mu           sync.Mutex
	pending      map[string]struct{}
	deleteBatchF func(keys []string)
	stopCh       chan struct{}
	flushMs      int64
}

func NewDeleteCoalescer(deleteBatch func(keys []string), flushMs int64) *DeleteCoalescer {
	if flushMs < 5 {
		flushMs = 5
	}
	if flushMs > 50 {
		flushMs = 50
	}

	return &DeleteCoalescer{
		pending:      make(map[string]struct{}),
		deleteBatchF: deleteBatch,
		stopCh:       make(chan struct{}),
		flushMs:      flushMs,
	}
}

func (dc *DeleteCoalescer) Enqueue(key string) {
	dc.mu.Lock()
	dc.pending[key] = struct{}{}
	dc.mu.Unlock()
}

func (dc *DeleteCoalescer) EnqueueBatch(keys []string) {
	if len(keys) == 0 {
		return
	}

	dc.mu.Lock()
	for _, k := range keys {
		dc.pending[k] = struct{}{}
	}
	dc.mu.Unlock()
}

func (dc *DeleteCoalescer) Start() {
	go func() {
		ticker := time.NewTicker(time.Duration(dc.flushMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dc.flush()
			case <-dc.stopCh:
				dc.flush()
				return
			}
		}
	}()
}

func (dc *DeleteCoalescer) flush() {
	dc.mu.Lock()
	if len(dc.pending) == 0 {
		dc.mu.Unlock()
		return
	}

	keys := make([]string, 0, len(dc.pending))
	for k := range dc.pending {
		keys = append(keys, k)
	}
	dc.pending = make(map[string]struct{})
	dc.mu.Unlock()

	if dc.deleteBatchF != nil {
		dc.deleteBatchF(keys)
	}
}

func (dc *DeleteCoalescer) Stop() {
	close(dc.stopCh)
}
