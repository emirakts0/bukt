package auth

import "sync"

var (
	instance     *TokenManager
	instanceOnce sync.Once
)

// Initialize sets up the singleton token manager
// Must be called once at application startup
func Initialize(secretKey []byte) {
	instanceOnce.Do(func() {
		instance = NewTokenManager(secretKey)
	})
}

// Manager returns the singleton token manager instance
// Panics if Initialize() was not called
func Manager() *TokenManager {
	if instance == nil {
		panic("auth: Manager() called before Initialize()")
	}
	return instance
}
