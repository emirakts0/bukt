package config

import (
	"crypto/rand"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	EnvTokenSecret        = "TOKEN_SECRET"
	EnvServerPort         = "SERVER_PORT"
	EnvLoggingEnvironment = "LOGGING_ENVIRONMENT"
	EnvLoggingLevel       = "LOGGING_LEVEL"
	EnvShardCount         = "SHARD_COUNT"
)

const (
	DefaultTokenSecret        = "" // Will be generated at startup if not provided
	DefaultServerPort         = 8080
	DefaultLoggingEnvironment = "production"
	DefaultLoggingLevel       = "info"
	DefaultShardCount         = 64
)

type Configuration struct {
	Auth    AuthConfig
	Server  ServerConfig
	Logging LoggingConfig
	Store   StoreConfig
}

type AuthConfig struct {
	TokenSecret []byte
}

type ServerConfig struct {
	Port int
}

type LoggingConfig struct {
	Environment string
	Level       string
}

type StoreConfig struct {
	ShardCount int
}

func NewConfig() *Configuration {
	tokenSecret := getEnv(EnvTokenSecret, DefaultTokenSecret)
	var tokenSecretBytes []byte

	if tokenSecret == "" {
		// Generate a random 32-byte secret at startup
		tokenSecretBytes = generateRandomSecret()
	} else {
		tokenSecretBytes = []byte(tokenSecret)
	}

	return &Configuration{
		Auth: AuthConfig{
			TokenSecret: tokenSecretBytes,
		},
		Server: ServerConfig{
			Port: getEnvAsInt(EnvServerPort, DefaultServerPort),
		},
		Logging: LoggingConfig{
			Environment: getEnv(EnvLoggingEnvironment, DefaultLoggingEnvironment),
			Level:       getEnv(EnvLoggingLevel, DefaultLoggingLevel),
		},
		Store: StoreConfig{
			ShardCount: getEnvAsInt(EnvShardCount, DefaultShardCount),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

// generateRandomSecret creates a cryptographically secure random 32-byte secret
func generateRandomSecret() []byte {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		slog.Error("Failed to generate random secret, using fallback", "error", err)
		// Fallback to a deterministic secret (not ideal but better than crashing)
		return []byte("fallback-secret-key-not-secure-please-set-TOKEN_SECRET")
	}
	slog.Info("Generated random token secret for this session")
	return secret
}
