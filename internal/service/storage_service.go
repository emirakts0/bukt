package service

import (
	"context"
	"errors"
	"key-value-store/internal/model"
	"sync"
	"time"
)

//todo: expire date geçen kayıtları düzenli silecek bir goroutine ekle
//todo: loglama ekle, uber-go/zap kullanabiliriz
//todo: password doğrulaması ve middleware ekle
//todo: store edilen mapi ayrıca bir pakette tutabiliriz, burası sadece işlem için kullanılır.
//todo: ...

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyExpired       = errors.New("key expired")
	ErrInvalidTTL       = errors.New("invalid TTL")
	ErrKeyAlreadyExists = errors.New("key already exists")
)

type StorageService struct {
	store map[string]model.StorageEntry
	mu    sync.RWMutex
}

func NewStorageService() *StorageService {
	return &StorageService{
		store: make(map[string]model.StorageEntry),
	}
}

func (s *StorageService) Set(ctx context.Context, key, value string, ttl int64) (model.StorageEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ttl <= 0 {
		return model.StorageEntry{}, ErrInvalidTTL
	}

	if _, exists := s.store[key]; exists {
		return model.StorageEntry{}, ErrKeyAlreadyExists
	}

	entry := model.StorageEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	s.store[key] = entry
	return entry, nil
}

func (s *StorageService) Get(ctx context.Context, key string) (model.StorageEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]
	if !exists {
		return model.StorageEntry{}, ErrKeyNotFound
	}

	if entry.IsExpired() {
		delete(s.store, key)
		return model.StorageEntry{}, ErrKeyExpired
	}

	return entry, nil
}

func (s *StorageService) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.store[key]; !exists {
		return ErrKeyNotFound
	}

	delete(s.store, key)
	return nil
}
