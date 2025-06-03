package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"key-value-store/internal/config"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrBucketNotFound      = errors.New("bucket not found")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrInvalidBucketName   = errors.New("invalid bucket name")
)

// Bucket represents a storage bucket with its own shard container
type Bucket struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	CreatedAt      time.Time       `json:"created_at"`
	ShardCount     int             `json:"shard_count"`
	KeyCount       int64           `json:"key_count"`
	MemoryUsage    int64           `json:"memory_usage"`
	ShardContainer *ShardContainer `json:"-"` // Direct pointer, not exposed in JSON
}

// BucketManager manages multiple buckets with direct shard container pointers
type BucketManager interface {
	// Bucket management
	CreateBucket(name, description string, shardCount int) (*Bucket, error)
	GetBucket(name string) (*Bucket, error)
	DeleteBucket(name string) error
	ListBuckets() []Bucket
	BucketExists(name string) bool

	// Store operations (for data access)
	GetBucketStore(bucketName string) (Store, error)

	// Lifecycle
	Shutdown()
}

type bucketManager struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
	cfg     *config.Configuration
}

func generateBucketID() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp-based ID
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(bytes)
}

func NewBucketManager() BucketManager {
	cfg := config.Config()
	manager := &bucketManager{
		buckets: make(map[string]*Bucket),
		cfg:     cfg,
	}

	// Create default bucket
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
		return nil, ErrInvalidBucketName
	}

	if shardCount <= 0 {
		shardCount = m.cfg.Store.ShardCount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[name]; exists {
		return nil, ErrBucketAlreadyExists
	}

	// Create shard container for this bucket
	shardContainer := NewShardContainer(shardCount)

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
		return nil, ErrBucketNotFound
	}

	// Create a copy with updated stats
	bucketCopy := *bucket
	bucketCopy.KeyCount = int64(len(bucket.ShardContainer.Keys()))
	bucketCopy.MemoryUsage = bucket.ShardContainer.GetMemoryUsage()

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
		return ErrBucketNotFound
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
		bucketCopy.MemoryUsage = bucket.ShardContainer.GetMemoryUsage()
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

func (m *bucketManager) GetBucketStore(bucketName string) (Store, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bucket, exists := m.buckets[bucketName]
	if !exists {
		return nil, ErrBucketNotFound
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
