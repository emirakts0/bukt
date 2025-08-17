package errs

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidTTL       = errors.New("invalid TTL")
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrKeyExpired       = errors.New("key expired")
	ErrMemoryLimit      = errors.New("memory limit exceeded")
	ErrCompression      = errors.New("compression error")
	ErrDeletion         = errors.New("deletion error")
)

var (
	ErrBucketNotFound      = errors.New("bucket not found")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrInvalidBucketName   = errors.New("invalid bucket name")
)

var (
	ErrInconsistentState = errors.New("inconsistent state detected")
	ErrDiskStoreFailed   = errors.New("disk store operation failed")
)
