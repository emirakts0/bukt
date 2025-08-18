package core

import (
	"errors"
	"key-value-store/internal/config"
	"key-value-store/internal/core/engine"
	"key-value-store/internal/errs"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Bucket struct {
	ID             string
	Name           string
	Description    string
	CreatedAt      time.Time
	ShardCount     int
	KeyCount       int64
	MemoryUsage    int64
	ShardContainer *engine.ShardContainer
}

// BucketManager manages multiple buckets with direct shard container pointers
type BucketManager interface {
	CreateBucket(name, description string, shardCount int) (*Bucket, error)
	GetBucket(name string) (*Bucket, error)
	DeleteBucket(name string) error
	ListBuckets() []Bucket
	BucketExists(name string) bool
	GetBucketStore(bucketName string) (engine.Store, error)
	Shutdown()
}

type bucketManager struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
	cfg     *config.Configuration
}

func generateBucketID() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func NewBucketManager(cfg *config.Configuration) BucketManager {
	manager := &bucketManager{
		buckets: make(map[string]*Bucket),
		cfg:     cfg,
	}

	defaultBucket, err := manager.CreateBucket("default", "Default bucket", cfg.Store.ShardCount)
	if err != nil {
		slog.Error("Failed to create default bucket", "error", err)
	} else {
		slog.Info("Created default bucket", "name", defaultBucket.Name, "id", defaultBucket.ID, "shard_count", defaultBucket.ShardCount)
	}

	return manager
}

func (m *bucketManager) CreateBucket(name, description string, shardCount int) (*Bucket, error) {
	slog.Debug("BucketManager: Creating bucket", "name", name, "shard_count", shardCount)

	if name == "" {
		return nil, errs.ErrInvalidBucketName
	}

	if shardCount <= 0 {
		shardCount = m.cfg.Store.ShardCount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[name]; exists {
		return nil, errs.ErrBucketAlreadyExists
	}

	shardContainer := engine.NewShardContainer(name, shardCount, m.cfg.Engine)

	bucket := &Bucket{
		ID:             generateBucketID(),
		Name:           name,
		Description:    description,
		CreatedAt:      time.Now(),
		ShardCount:     shardCount,
		KeyCount:       0,
		MemoryUsage:    0,
		ShardContainer: shardContainer,
	}

	m.buckets[name] = bucket

	slog.Info("BucketManager: Created bucket", "name", name, "id", bucket.ID, "shard_count", shardCount)
	return bucket, nil
}

func (m *bucketManager) GetBucket(name string) (*Bucket, error) {
	slog.Debug("BucketManager: Getting bucket", "name", name)

	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, exists := m.buckets[name]
	if !exists {
		return nil, errs.ErrBucketNotFound
	}

	// Create a copy with updated stats
	bucketCopy := *bucket
	bucketCopy.KeyCount = int64(len(bucket.ShardContainer.Keys()))
	bucketCopy.MemoryUsage = bucket.ShardContainer.Usage()

	return &bucketCopy, nil
}

func (m *bucketManager) DeleteBucket(name string) error {
	slog.Debug("BucketManager: Deleting bucket", "name", name)

	// Prevent deletion of default bucket
	if name == "default" {
		return errors.New("cannot delete default bucket")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, exists := m.buckets[name]
	if !exists {
		return errs.ErrBucketNotFound
	}

	// Stop GC for the shard container
	bucket.ShardContainer.StopGC()

	// Delete the bucket
	delete(m.buckets, name)

	slog.Info("BucketManager: Deleted bucket", "name", name, "id", bucket.ID)
	return nil
}

func (m *bucketManager) ListBuckets() []Bucket {
	slog.Debug("BucketManager: Listing buckets")

	m.mu.RLock()
	defer m.mu.RUnlock()

	buckets := make([]Bucket, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		bucketCopy := *bucket // Create a copy
		bucketCopy.KeyCount = int64(len(bucket.ShardContainer.Keys()))
		bucketCopy.MemoryUsage = bucket.ShardContainer.Usage()
		buckets = append(buckets, bucketCopy)
	}

	return buckets
}

func (m *bucketManager) BucketExists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.buckets[name]
	return exists
}

func (m *bucketManager) GetBucketStore(bucketName string) (engine.Store, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, exists := m.buckets[bucketName]
	if !exists {
		return nil, errs.ErrBucketNotFound
	}

	return bucket.ShardContainer, nil
}

func (m *bucketManager) Shutdown() {
	slog.Info("BucketManager: Shutting down")

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, bucket := range m.buckets {
		slog.Debug("BucketManager: Stopping GC for bucket", "name", name, "id", bucket.ID)
		bucket.ShardContainer.StopGC()
	}
}
