package gc

import (
	"math/rand"
	"sync"
	"time"
)

type wheelEntry struct {
	key      string
	expireAt int64
	rounds   int
}

type wheelSlot struct {
	entries []wheelEntry
}

type HashedWheel struct {
	slots      []wheelSlot
	slotCount  int
	tickMs     int64
	currentPos int
	startTime  int64
	mu         sync.Mutex
	keyIndex   map[string]int
}

func NewHashedWheel(slotCount int, tickMs int64) *HashedWheel {
	if slotCount < 64 {
		slotCount = 64
	}
	if slotCount > 8192 {
		slotCount = 8192
	}
	if tickMs < 100 {
		tickMs = 100
	}
	if tickMs > 1000 {
		tickMs = 1000
	}

	return &HashedWheel{
		slots:      make([]wheelSlot, slotCount),
		slotCount:  slotCount,
		tickMs:     tickMs,
		currentPos: 0,
		startTime:  time.Now().UnixNano() / 1e6,
		keyIndex:   make(map[string]int),
	}
}

func (hw *HashedWheel) Add(key string, expireAt int64) {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	smear := rand.Int63n(hw.tickMs / 4)
	if rand.Intn(2) == 0 {
		smear = -smear
	}
	expireAt += smear

	now := time.Now().UnixNano() / 1e6
	delay := expireAt/1e6 - now

	if delay < 0 {
		delay = 0
	}

	ticks := delay / hw.tickMs
	slot := (hw.currentPos + int(ticks)) % hw.slotCount
	rounds := int(ticks) / hw.slotCount

	hw.removeLocked(key)

	hw.slots[slot].entries = append(hw.slots[slot].entries, wheelEntry{
		key:      key,
		expireAt: expireAt,
		rounds:   rounds,
	})
	hw.keyIndex[key] = slot
}

func (hw *HashedWheel) Remove(key string) {
	hw.mu.Lock()
	defer hw.mu.Unlock()
	hw.removeLocked(key)
}

func (hw *HashedWheel) removeLocked(key string) {
	slot, exists := hw.keyIndex[key]
	if !exists {
		return
	}

	entries := hw.slots[slot].entries
	for i := 0; i < len(entries); i++ {
		if entries[i].key == key {
			hw.slots[slot].entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	delete(hw.keyIndex, key)
}

func (hw *HashedWheel) Advance() []string {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	now := time.Now().UnixNano() / 1e6
	slot := &hw.slots[hw.currentPos]
	var expired []string
	remaining := []wheelEntry{}

	for i := range slot.entries {
		e := &slot.entries[i]
		if e.rounds > 0 {
			e.rounds--
			remaining = append(remaining, *e)
		} else if e.expireAt/1e6 <= now {
			expired = append(expired, e.key)
			delete(hw.keyIndex, e.key)
		} else {
			remaining = append(remaining, *e)
		}
	}

	slot.entries = remaining
	hw.currentPos = (hw.currentPos + 1) % hw.slotCount
	return expired
}
