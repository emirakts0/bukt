package bucket

import (
	"crypto/subtle"
	"encoding/hex"
	"key-value-store/internal/config"
	"key-value-store/internal/engine"
	"key-value-store/internal/errs"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type BucketMetadata struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	ShardCount  int
	KeyCount    int64
	MemoryUsage int64
	authToken   [16]byte
	store       *engine.ShardContainer
}

type BucketIndex struct {
	buckets map[string]*BucketMetadata
}

type BucketManager interface {
	CreateBucket(name, description string, shardCount int) (string, error)
	GetBucket(name string) (*BucketMetadata, bool)
	DeleteBucket(name string, token []byte) error
	ListBuckets() []*BucketMetadata
	BucketExists(name string) bool
	AuthenticateAndGetStore(name string, token []byte) (*engine.ShardContainer, bool)
	Shutdown()
}

type bucketManager struct {
	ptr     atomic.Pointer[BucketIndex]
	writeMu sync.Mutex
	cfg     *config.Configuration
}

func NewBucketManager(cfg *config.Configuration) BucketManager {
	bm := &bucketManager{cfg: cfg}
	idx := &BucketIndex{buckets: make(map[string]*BucketMetadata)}
	bm.ptr.Store(idx)

	token, err := bm.CreateBucket("default", "Default bucket", cfg.Store.ShardCount)
	if err != nil {
		slog.Error("Failed to create default bucket", "error", err)
	} else {
		slog.Info("Created default bucket", "token", token, "shard_count", cfg.Store.ShardCount)
	}

	return bm
}

func (bm *bucketManager) snapshot() *BucketIndex {
	return bm.ptr.Load()
}

func (bm *bucketManager) GetBucket(name string) (*BucketMetadata, bool) {
	idx := bm.snapshot()
	b, ok := idx.buckets[name]
	if !ok {
		return nil, false
	}

	meta := &BucketMetadata{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		ShardCount:  b.ShardCount,
	}
	if b.store != nil {
		meta.KeyCount = b.store.Count()
		meta.MemoryUsage = b.store.Usage()
	}
	return meta, true
}

func (bm *bucketManager) AuthenticateAndGetStore(name string, token []byte) (*engine.ShardContainer, bool) {
	if len(token) != 16 {
		return nil, false
	}

	idx := bm.snapshot()
	b, ok := idx.buckets[name]
	if !ok {
		return nil, false
	}

	if subtle.ConstantTimeCompare(b.authToken[:], token) != 1 {
		return nil, false
	}

	return b.store, true
}

func (bm *bucketManager) CreateBucket(name, description string, shardCount int) (string, error) {
	if name == "" {
		return "", errs.ErrInvalidBucketName
	}

	if shardCount <= 0 {
		shardCount = bm.cfg.Store.ShardCount
	}

	tokenBytes := generateAuthToken()

	bm.writeMu.Lock()
	defer bm.writeMu.Unlock()

	old := bm.snapshot()
	if _, exists := old.buckets[name]; exists {
		return "", errs.ErrBucketAlreadyExists
	}

	newIdx := &BucketIndex{
		buckets: make(map[string]*BucketMetadata, len(old.buckets)+1),
	}

	for k, v := range old.buckets {
		newIdx.buckets[k] = v
	}

	shardContainer := engine.NewShardContainer(shardCount)

	meta := &BucketMetadata{
		ID:          generateBucketID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		ShardCount:  shardCount,
		authToken:   tokenBytes,
		store:       shardContainer,
	}

	newIdx.buckets[name] = meta
	bm.ptr.Store(newIdx)

	tokenStr := hex.EncodeToString(tokenBytes[:])
	slog.Info("BucketManager: Created bucket", "name", name, "id", meta.ID, "shard_count", shardCount)
	return tokenStr, nil
}

func (bm *bucketManager) DeleteBucket(name string, token []byte) error {
	if name == "default" {
		return errs.ErrCannotDeleteDefault
	}

	if len(token) != 16 {
		return errs.ErrUnauthorized
	}

	bm.writeMu.Lock()
	defer bm.writeMu.Unlock()

	old := bm.snapshot()
	b, exists := old.buckets[name]
	if !exists {
		return errs.ErrBucketNotFound
	}

	if subtle.ConstantTimeCompare(b.authToken[:], token) != 1 {
		return errs.ErrUnauthorized
	}

	newIdx := &BucketIndex{
		buckets: make(map[string]*BucketMetadata, len(old.buckets)-1),
	}

	for k, v := range old.buckets {
		if k != name {
			newIdx.buckets[k] = v
		}
	}

	bm.ptr.Store(newIdx)

	if b.store != nil {
		b.store.StopGC()
	}

	slog.Info("BucketManager: Deleted bucket", "name", name, "id", b.ID)
	return nil
}

func (bm *bucketManager) ListBuckets() []*BucketMetadata {
	idx := bm.snapshot()
	result := make([]*BucketMetadata, 0, len(idx.buckets))

	for _, b := range idx.buckets {
		meta := &BucketMetadata{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			CreatedAt:   b.CreatedAt,
			ShardCount:  b.ShardCount,
		}
		if b.store != nil {
			meta.KeyCount = b.store.Count()
			meta.MemoryUsage = b.store.Usage()
		}
		result = append(result, meta)
	}

	return result
}

func (bm *bucketManager) BucketExists(name string) bool {
	idx := bm.snapshot()
	_, exists := idx.buckets[name]
	return exists
}

func (bm *bucketManager) Shutdown() {
	slog.Info("BucketManager: Shutting down")

	bm.writeMu.Lock()
	defer bm.writeMu.Unlock()

	idx := bm.snapshot()
	for name, b := range idx.buckets {
		slog.Debug("BucketManager: Closing bucket", "name", name, "id", b.ID)
		if b.store != nil {
			if err := b.store.Close(); err != nil {
				slog.Error("BucketManager: Failed to close bucket", "name", name, "error", err)
			}
		}
	}
}

func generateAuthToken() [16]byte {
	id, _ := uuid.NewV7()
	return id
}

func generateBucketID() string {
	id, _ := uuid.NewV7()
	return id.String()
}
