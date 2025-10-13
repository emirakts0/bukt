package engine

import (
	"key-value-store/internal/gc"
	"key-value-store/internal/util"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type entry = StorageEntry

type Segment struct {
	keys []string
	vals []*entry
}

type Index struct {
	segments []Segment
}

type COWIndexStore struct {
	ptr       atomic.Pointer[Index]
	writeMu   sync.Mutex
	usedBytes int64
	keyCount  int64
	hasher    util.Hasher
	*gc.GarbageCollector
}

func NewMemoryStore() Store {
	idx := &Index{segments: make([]Segment, 64)}
	s := &COWIndexStore{
		hasher: util.NewDefaultHasher(),
	}
	s.ptr.Store(idx)
	s.GarbageCollector = gc.NewGarbageCollector(s.deleteBatch)
	return s
}

func (s *COWIndexStore) snapshot() *Index { return s.ptr.Load() }

func (s *COWIndexStore) segmentIndexOf(key string) int {
	t := s.ptr.Load()
	hash := s.hasher.Sum64String(key)
	return int(hash % uint64(len(t.segments)))
}

func (s *COWIndexStore) Get(key string) (StorageEntry, bool) {
	idx := s.snapshot()
	seg := &idx.segments[s.segmentIndexOf(key)]

	i, found := slices.BinarySearch(seg.keys, key)
	if !found {
		return StorageEntry{}, false
	}
	e := seg.vals[i]

	if e.IsExpired() {
		s.GarbageCollector.ScheduleDelete(key)
		return StorageEntry{}, false
	}
	if e.SingleRead {
		if atomic.CompareAndSwapInt32(&e.AccessCount, 0, 1) {
			s.GarbageCollector.ScheduleDelete(key)
			return *e, true
		}
		return StorageEntry{}, false
	}

	// Update access stats with cached time (no syscall overhead)
	atomic.AddInt32(&e.AccessCount, 1)
	atomic.StoreInt64(&e.LastAccess, util.CachedNow())

	return *e, true
}

func (s *COWIndexStore) Set(key string, val StorageEntry) {
	val.LastAccess = util.CachedNow()
	val.AccessCount = 0

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	old := s.snapshot()
	n := len(old.segments)

	newIdx := &Index{segments: make([]Segment, n)}
	copy(newIdx.segments, old.segments)

	si := s.segmentIndexOf(key)
	oldSeg := old.segments[si]

	ns := Segment{
		keys: append([]string(nil), oldSeg.keys...),
		vals: append([]*entry(nil), oldSeg.vals...),
	}

	i, found := slices.BinarySearch(ns.keys, key)
	var delta int64
	if found {
		prev := ns.vals[i]
		delta -= sizeOf(key, prev)
		ns.vals[i] = &val
	} else {
		ns.keys = append(ns.keys, "")
		copy(ns.keys[i+1:], ns.keys[i:])
		ns.keys[i] = key

		ns.vals = append(ns.vals, nil)
		copy(ns.vals[i+1:], ns.vals[i:])
		ns.vals[i] = &val
		atomic.AddInt64(&s.keyCount, 1)
	}
	delta += sizeOf(key, &val)

	newIdx.segments[si] = ns
	s.ptr.Store(newIdx)
	atomic.AddInt64(&s.usedBytes, delta)

	s.GarbageCollector.Schedule(key, val.TTL, val.CreatedAt)
}

func (s *COWIndexStore) Delete(key string) { s.deleteBatch([]string{key}) }

func (s *COWIndexStore) deleteBatch(keys []string) {
	if len(keys) == 0 {
		return
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	old := s.snapshot()
	n := len(old.segments)
	newIdx := &Index{segments: make([]Segment, n)}
	copy(newIdx.segments, old.segments)

	// Pre-allocate map with capacity hint to reduce allocations
	group := make(map[int][]string, len(keys)/4+1)
	for _, k := range keys {
		si := s.segmentIndexOf(k)
		group[si] = append(group[si], k)
	}

	var delta int64
	for si, gks := range group {
		oldSeg := old.segments[si]
		ns := Segment{
			keys: append([]string(nil), oldSeg.keys...),
			vals: append([]*entry(nil), oldSeg.vals...),
		}
		for _, k := range gks {
			i, found := slices.BinarySearch(ns.keys, k)
			if !found {
				continue
			}
			ent := ns.vals[i]
			delta -= sizeOf(k, ent)
			ns.keys = append(ns.keys[:i], ns.keys[i+1:]...)
			ns.vals = append(ns.vals[:i], ns.vals[i+1:]...)
			atomic.AddInt64(&s.keyCount, -1)
			s.GarbageCollector.Cancel(k)
		}
		newIdx.segments[si] = ns
	}

	if delta != 0 {
		s.ptr.Store(newIdx)
		atomic.AddInt64(&s.usedBytes, delta)
	}
}

func (s *COWIndexStore) Exists(key string) bool {
	idx := s.snapshot()
	seg := &idx.segments[s.segmentIndexOf(key)]
	_, found := slices.BinarySearch(seg.keys, key)
	return found
}

func (s *COWIndexStore) Keys() []string {
	idx := s.snapshot()
	var total int
	for i := range idx.segments {
		total += len(idx.segments[i].keys)
	}
	out := make([]string, 0, total)
	for i := range idx.segments {
		out = append(out, idx.segments[i].keys...)
	}
	return out
}

func (s *COWIndexStore) Usage() int64 { return atomic.LoadInt64(&s.usedBytes) }

func (s *COWIndexStore) Count() int64 { return atomic.LoadInt64(&s.keyCount) }

func sizeOf(key string, e *StorageEntry) int64 {
	var v int64
	if e.OriginalSize > 0 {
		v = e.OriginalSize
	} else {
		v = int64(len(e.Value))
	}
	return int64(len(key)) + v
}

func (s *COWIndexStore) StartGC(d time.Duration) { s.GarbageCollector.Start(d) }
func (s *COWIndexStore) StopGC()                 { s.GarbageCollector.Stop() }
