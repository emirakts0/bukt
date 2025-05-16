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
