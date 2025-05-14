package service

import (
	"context"
	"errors"
	"key-value-store/internal/model"
	"key-value-store/internal/store"
	"log/slog"
	"time"
)

//todo: expire date geçen kayıtları düzenli silecek bir goroutine ekle
//todo: value değeri sadece string değil, int, float, boolean vs. olabilir.
//todo: requestleri takip etmek için logda correlation id görelim.
//todo: performansı test et, ekstra depolama özellikleri ekle. Veri kaybının önlenmesi vb için. tek mapde tutuluyor her şey cok yetersiz ve basit şuanda.
//todo: dosya isimlerini metot ve değişken isimlerini daha iyi design et, web servis havasını bir nebze azalt.
//todo: healthcheck endpointi ekle, depolama bilgileri vb vb gibi şeyler için.
//todo: birden fazla faklı memory storelar olabilir. ayrıca bunlara queue özelliği de ekle, istenirse queue olarak kullanılabilsin.
//todo: auth kısmını da bunlara göre yenilemek gerek, hangi kovaya kim erişebilir konusunu netleştirmek için.

const (
	GCInterval = 60
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidTTL       = errors.New("invalid TTL")
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrKeyExpired       = errors.New("key expired")
)

type IStorageService interface {
	Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error)
	Get(ctx context.Context, key string) (model.StorageEntry, error)
	Delete(ctx context.Context, key string) error
}

type storageService struct {
	store store.MemoryStore
}

func NewStorageService() IStorageService {
	s := &storageService{
		store: store.NewMemoryStore(),
	}
	s.StartGC(GCInterval * time.Second)
	return s
}

func (s *storageService) Set(ctx context.Context, key, value string, ttl int64, singleRead bool) (model.StorageEntry, error) {
	slog.Debug("Attempting to set key-value pair", "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl <= 0 {
		slog.Warn("Invalid TTL provided", "key", key, "ttl", ttl)
		return model.StorageEntry{}, ErrInvalidTTL
	}

	if s.store.Exists(key) {
		slog.Warn("Key already exists", "key", key)
		return model.StorageEntry{}, ErrKeyAlreadyExists
	}

	now := time.Now()
	entry := model.StorageEntry{
		Key:        key,
		Value:      value,
		TTL:        ttl,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
		SingleRead: singleRead,
	}

	s.store.Set(key, entry)
	slog.Info("Successfully set key-value pair", "key", key, "ttl", entry.TTL, "single_read", singleRead)
	return entry, nil
}

func (s *storageService) Get(ctx context.Context, key string) (model.StorageEntry, error) {
	slog.Debug("Attempting to get value", "key", key)

	entry, exists := s.store.Get(key)
	if !exists {
		slog.Warn("Key not found", "key", key)
		return model.StorageEntry{}, ErrKeyNotFound
	}

	if entry.IsExpired() {
		slog.Warn("Key has expired", "key", key, "expires_at", entry.ExpiresAt.Format(time.RFC3339))
		s.store.Delete(key)
		return model.StorageEntry{}, ErrKeyExpired
	}

	// If the entry is single-read, delete it after reading
	if entry.SingleRead {
		s.store.Delete(key)
		slog.Info("Deleted single-read key after reading", "key", key, "single_read", entry.SingleRead)
	}

	slog.Info("Successfully retrieved value", "key", key, "single_read", entry.SingleRead)
	return entry, nil
}

func (s *storageService) Delete(ctx context.Context, key string) error {
	slog.Debug("Attempting to delete key", "key", key)

	if !s.store.Delete(key) {
		slog.Warn("Key not found for deletion", "key", key)
		return ErrKeyNotFound
	}

	slog.Info("Successfully deleted key", "key", key)
	return nil
}

func (s *storageService) StartGC(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				slog.Debug("Running garbage collection for expired keys.")
				s.cleanupExpiredKeys()
			}
		}
	}()
}

func (s *storageService) cleanupExpiredKeys() {
	keys := s.store.Keys()

	for _, key := range keys {
		entry, exists := s.store.Get(key)
		if exists && entry.IsExpired() {
			s.store.Delete(key)
			slog.Info("Expired key deleted by GC", "key", key, "expired_at", entry.ExpiresAt.Format(time.RFC3339))
		}
	}
}
