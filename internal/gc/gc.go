package gc

import (
	"sync"
	"time"
)

const (
	defaultWheelSlots   = 2048
	defaultWheelTickMs  = 500
	defaultNearWindowMs = 2000
	defaultCoalescerMs  = 10
	defaultNearTickerMs = 50
)

type GarbageCollector struct {
	wheel     *HashedWheel
	nearQueue *NearQueue
	coalescer *DeleteCoalescer
	stopCh    chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	tickMs    int64
}

func NewGarbageCollector(deleteBatch func(keys []string)) *GarbageCollector {
	return &GarbageCollector{
		wheel:     NewHashedWheel(defaultWheelSlots, defaultWheelTickMs),
		nearQueue: NewNearQueue(defaultNearWindowMs),
		coalescer: NewDeleteCoalescer(deleteBatch, defaultCoalescerMs),
		stopCh:    make(chan struct{}),
		tickMs:    defaultWheelTickMs,
	}
}

func (gc *GarbageCollector) Schedule(key string, ttl int64, createdAt time.Time) {
	if ttl <= 0 {
		gc.Cancel(key)
		return
	}

	expireAt := createdAt.Add(time.Duration(ttl)).UnixNano()

	if !gc.nearQueue.Add(key, expireAt) {
		gc.wheel.Add(key, expireAt)
	}
}

func (gc *GarbageCollector) Cancel(key string) {
	gc.nearQueue.Remove(key)
	gc.wheel.Remove(key)
}

func (gc *GarbageCollector) ScheduleDelete(key string) {
	gc.coalescer.Enqueue(key)
}

func (gc *GarbageCollector) Start(interval time.Duration) {
	gc.startOnce.Do(func() {
		gc.coalescer.Start()

		go func() {
			nearTicker := time.NewTicker(defaultNearTickerMs * time.Millisecond)
			wheelTicker := time.NewTicker(time.Duration(gc.tickMs) * time.Millisecond)
			defer nearTicker.Stop()
			defer wheelTicker.Stop()

			for {
				select {
				case <-nearTicker.C:
					expired := gc.nearQueue.DrainExpired()
					gc.coalescer.EnqueueBatch(expired)
				case <-wheelTicker.C:
					expired := gc.wheel.Advance()
					gc.coalescer.EnqueueBatch(expired)
				case <-gc.stopCh:
					return
				}
			}
		}()
	})
}

func (gc *GarbageCollector) Stop() {
	gc.stopOnce.Do(func() {
		close(gc.stopCh)
		gc.coalescer.Stop()
	})
}
