package service

import (
	"errors"
	"key-value-store/internal/logger"
	"key-value-store/internal/model"
	"key-value-store/internal/store"
	"time"

	"go.uber.org/zap"
)

//todo: sadece standard library kullanacak biçime geçirelim.
//todo: expire date geçen kayıtları düzenli silecek bir goroutine ekle
//todo: value değeri sadece string değil, int, float, boolean vs. olabilir.
//todo: requestleri takip etmek için logda correlation id görelim.
//todo: performansı test et, ekstra depolama özellikleri ekle. Veri kaybının önlenmesi vb için. tek mapde tutuluyor her şey cok yetersiz ve basit şuanda.
//todo: ayrıca servis işlem kısmı ile handlerları daha keskin bir biçimde ayır, loglamaları da ona göre ayarla. daha sonradan farklı kanallardan erişim eklendiğinde aynı kod base kullanılacak.
//todo: dosya isimlerini metot ve değişken isimlerini daha iyi design et, web servis havasını bir nebze azalt.
//todo: healthcheck endpointi ekle, depolama bilgileri vb vb gibi şeyler için.
//todo: birden fazla faklı memory storelar olabilir. ayrıca bunlara queue özelliği de ekle, istenirse queue olarak kullanılabilsin.
//todo: auth kısmını da bunlara göre yenilemek gerek, hangi kovaya kim erişebilir konusunu netleştirmek için.
//todo: ...

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidTTL       = errors.New("invalid TTL")
	ErrKeyAlreadyExists = errors.New("key already exists")
)

type IStorageService interface {
	Set(key, value string, ttl int64, singleRead bool) (model.StorageEntry, error)
	Get(key string) (model.StorageEntry, error)
	Delete(key string) error
}

type storageService struct {
	store store.MemoryStore
	log   *zap.SugaredLogger
}

func NewStorageService() IStorageService {
	return &storageService{
		store: store.NewMemoryStore(),
		log:   logger.GetSugared(),
	}
}

func (s *storageService) Set(key, value string, ttl int64, singleRead bool) (model.StorageEntry, error) {
	s.log.Debugw("Attempting to set key-value pair", "key", key, "ttl", ttl, "single_read", singleRead)

	if ttl <= 0 {
		s.log.Warnw("Invalid TTL provided", "key", key, "ttl", ttl)
		return model.StorageEntry{}, ErrInvalidTTL
	}

	if s.store.Exists(key) {
		s.log.Warnw("Key already exists", "key", key)
		return model.StorageEntry{}, ErrKeyAlreadyExists
	}

	now := time.Now()
	entry := model.StorageEntry{
		Key:        key,
		Value:      value,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
		SingleRead: singleRead,
	}

	s.store.Set(key, entry)
	s.log.Infow("Successfully set key-value pair", "key", key, "expires_at", entry.ExpiresAt, "single_read", singleRead)
	return entry, nil
}

func (s *storageService) Get(key string) (model.StorageEntry, error) {
	s.log.Debugw("Attempting to get value", "key", key)

	entry, exists := s.store.Get(key)
	if !exists {
		s.log.Warnw("Key not found", "key", key)
		return model.StorageEntry{}, ErrKeyNotFound
	}

	if entry.IsExpired() {
		s.log.Warnw("Key has expired", "key", key, "expires_at", entry.ExpiresAt)
		return model.StorageEntry{}, ErrKeyNotFound
	}

	// If the entry is single-read, delete it after reading
	if entry.SingleRead {
		s.store.Delete(key)
		s.log.Infow("Deleted single-read key after reading", "key", key)
	}

	s.log.Infow("Successfully retrieved value", "key", key, "expires_at", entry.ExpiresAt, "single_read", entry.SingleRead)
	return entry, nil
}

func (s *storageService) Delete(key string) error {
	s.log.Debugw("Attempting to delete key", "key", key)

	if !s.store.Delete(key) {
		s.log.Warnw("Key not found for deletion", "key", key)
		return ErrKeyNotFound
	}

	s.log.Infow("Successfully deleted key", "key", key)
	return nil
}
