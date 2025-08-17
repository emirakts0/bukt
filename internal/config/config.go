package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
	EnvEngineType           = "ENGINE_TYPE"
	EnvEngineDataDir        = "ENGINE_DATA_DIR"
	EnvEngineEvictionInt    = "ENGINE_EVICTION_INTERVAL"
	EnvEngineEvictionBatch  = "ENGINE_EVICTION_BATCH_SIZE"
	EnvEngineCompactionInt  = "ENGINE_COMPACTION_INTERVAL"
)

const (
	// Default values
	DefaultAuthUsername         = "emir"
	DefaultAuthPassword         = "emir"
	DefaultServerPort           = 8080
	DefaultLoggingEnvironment   = "production"
	DefaultLoggingLevel         = "info"
	DefaultShardCount           = 4
	DefaultCompressionType      = "none"
	DefaultCompressionThreshold = 0           // 1024 1KB
	DefaultEngineType           = "in-memory" // in-memory, tiered
	DefaultEngineDataDir        = "data"
	DefaultEngineEvictionInt    = 1 * time.Minute
	DefaultEngineEvictionBatch  = 100
	DefaultEngineCompactionInt  = 5 * time.Minute
)

var (
	config     *Configuration
	configOnce sync.Once
)

type Configuration struct {
	Auth    AuthConfig
	Server  ServerConfig
	Logging LoggingConfig
	Store   StoreConfig
	Engine  EngineConfig
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
	ShardCount           int
	CompressionType      string
	CompressionThreshold int64
}

type EngineConfig struct {
	Type               string
	DataDir            string
	EvictionInterval   time.Duration
	EvictionBatchSize  int
	CompactionInterval time.Duration
}

func Config() *Configuration {
	configOnce.Do(func() {
		config = &Configuration{
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
			Engine: EngineConfig{
				Type:               getEnv(EnvEngineType, DefaultEngineType),
				DataDir:            getEnv(EnvEngineDataDir, DefaultEngineDataDir),
				EvictionInterval:   getEnvAsDuration(EnvEngineEvictionInt, DefaultEngineEvictionInt),
				EvictionBatchSize:  getEnvAsInt(EnvEngineEvictionBatch, DefaultEngineEvictionBatch),
				CompactionInterval: getEnvAsDuration(EnvEngineCompactionInt, DefaultEngineCompactionInt),
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

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
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
