package engine

import (
	"encoding/json"
	"key-value-store/internal/core"
	"log/slog"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type DiskStore struct {
	db *badger.DB
	*GarbageCollector
}

func NewDiskStore(filePath string) (Store, error) {
	opts := badger.DefaultOptions(filePath).
		WithLogger(nil).
		WithBlockCacheSize(0).
		WithIndexCacheSize(0)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	ds := &DiskStore{
		db: db,
	}
	ds.GarbageCollector = NewGarbageCollector(ds.Delete)

	go ds.runValueLogGC()

	return ds, nil
}

func (s *DiskStore) Set(key string, entry core.StorageEntry) {
	err := s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		e := badger.NewEntry([]byte(key), data)
		if entry.TTL > 0 {
			e = e.WithTTL(time.Duration(entry.TTL))
		}

		return txn.SetEntry(e)
	})

	if err != nil {
		slog.Error("DiskStore: Failed to set key", "key", key, "error", err)
		return
	}

	s.GarbageCollector.Track(key, &entry)
}

func (s *DiskStore) Get(key string) (core.StorageEntry, bool) {
	var entry core.StorageEntry

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &entry)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return core.StorageEntry{}, false
		}
		slog.Error("DiskStore: Failed to get key", "key", key, "error", err)
		return core.StorageEntry{}, false
	}

	if entry.IsExpired() {
		go s.Delete(key)
		return core.StorageEntry{}, false
	}

	return entry, true
}

func (s *DiskStore) Delete(key string) {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil && err != badger.ErrKeyNotFound {
		slog.Error("DiskStore: Failed to delete key", "key", key, "error", err)
	}

	s.GarbageCollector.Untrack(key)
}

func (s *DiskStore) Exists(key string) bool {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	return err == nil
}

func (s *DiskStore) Keys() []string {
	var keys []string

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			keys = append(keys, key)
		}

		return nil
	})

	if err != nil {
		slog.Error("DiskStore: Failed to get keys", "error", err)
		return nil
	}

	return keys
}

func (s *DiskStore) StartGC(interval time.Duration) {
	s.GarbageCollector.Start(interval)
}

func (s *DiskStore) StopGC() {
	s.GarbageCollector.Stop()
}

func (s *DiskStore) runValueLogGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		err := s.db.RunValueLogGC(0.5)
		if err != nil && err != badger.ErrNoRewrite {
			slog.Debug("DiskStore: Value log GC completed", "error", err)
		}
	}
}

func (s *DiskStore) Usage() int64 {
	lsm, vlog := s.db.Size()
	return lsm + vlog
}

func (s *DiskStore) Close() error {
	s.StopGC()
	return s.db.Close()
}
