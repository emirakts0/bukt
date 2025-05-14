package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	// Environment variable keys
	EnvAuthUsername       = "AUTH_USERNAME"
	EnvAuthPassword       = "AUTH_PASSWORD"
	EnvServerPort         = "SERVER_PORT"
	EnvLoggingEnvironment = "LOGGING_ENVIRONMENT"
	EnvLoggingLevel       = "LOGGING_LEVEL"

	// Default values
	DefaultAuthUsername       = "emir"
	DefaultAuthPassword       = "emir"
	DefaultServerPort         = 8080
	DefaultLoggingEnvironment = "development"
	DefaultLoggingLevel       = "debug"
)

var (
	config     *Config
	configOnce sync.Once
)

type Config struct {
	Auth    AuthConfig
	Server  ServerConfig
	Logging LoggingConfig
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

func Get() *Config {
	configOnce.Do(func() {
		config = &Config{
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
