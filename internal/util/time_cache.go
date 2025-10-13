package util

import (
	"sync/atomic"
	"time"
)

// Global cached time (updated every 10ms)
var cachedTimeNanos atomic.Int64

func init() {
	cachedTimeNanos.Store(time.Now().UnixNano())

	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			cachedTimeNanos.Store(time.Now().UnixNano())
		}
	}()
}

func CachedNow() int64 {
	return cachedTimeNanos.Load()
}
