package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	// Environment variable keys
	EnvAuthUsername         = "AUTH_USERNAME"
	EnvAuthPassword         = "AUTH_PASSWORD"
	EnvServerPort           = "SERVER_PORT"
	EnvLoggingEnvironment   = "LOGGING_ENVIRONMENT"
	EnvLoggingLevel         = "LOGGING_LEVEL"
	EnvShardCount           = "SHARD_COUNT"
	EnvCompressionType      = "COMPRESSION_TYPE"
	EnvCompressionThreshold = "COMPRESSION_THRESHOLD"

	// Default values
	DefaultAuthUsername         = "emir"
	DefaultAuthPassword         = "emir"
	DefaultServerPort           = 8080
	DefaultLoggingEnvironment   = "production"
	DefaultLoggingLevel         = "info"
	DefaultShardCount           = 4
	DefaultCompressionType      = "none"
	DefaultCompressionThreshold = 0 // 1024 1KB
)

var (
	config     *Configs
	configOnce sync.Once
)

type Configs struct {
	Auth    AuthConfig
	Server  ServerConfig
	Logging LoggingConfig
	Store   StoreConfig
}

type AuthConfig struct {
	Username string
	Password string
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

	EvictionPolicy       string
	CompressionType      string
	CompressionThreshold int64
}

func Get() *Configs {
	configOnce.Do(func() {
		config = &Configs{
			Auth: AuthConfig{
				Username: getEnv(EnvAuthUsername, DefaultAuthUsername),
				Password: getEnv(EnvAuthPassword, DefaultAuthPassword),
			},
			Server: ServerConfig{
				Port: getEnvAsInt(EnvServerPort, DefaultServerPort),
			},
			Logging: LoggingConfig{
				Environment: getEnv(EnvLoggingEnvironment, DefaultLoggingEnvironment),
				Level:       getEnv(EnvLoggingLevel, DefaultLoggingLevel),
			},
			Store: StoreConfig{
				ShardCount:           getEnvAsInt(EnvShardCount, DefaultShardCount),
				CompressionType:      getEnv(EnvCompressionType, DefaultCompressionType),
				CompressionThreshold: getEnvAsInt64(EnvCompressionThreshold, DefaultCompressionThreshold),
			},
		}
	})
	return config
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
