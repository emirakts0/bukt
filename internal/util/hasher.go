package util

import "github.com/cespare/xxhash/v2"

func NewDefaultHasher() Hasher {
	return xxhasher{}
}

type xxhasher struct{}

func (x xxhasher) Sum64(key []byte) uint64 {
	return xxhash.Sum64(key)
}

func (x xxhasher) Sum64String(key string) uint64 {
	return xxhash.Sum64String(key)
}

type Hasher interface {
	Sum64([]byte) uint64
	Sum64String(string) uint64
}
