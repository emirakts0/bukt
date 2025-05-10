package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

var (
	config      *Config
	configOnce  sync.Once
	defaultPath = "./configs/config.yaml"
)

type Config struct {
	Auth    AuthConfig    `mapstructure:"auth"`
	Server  ServerConfig  `mapstructure:"server"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type AuthConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type LoggingConfig struct {
	Environment string `mapstructure:"environment"`
	Level       string `mapstructure:"level"`
}

// Get returns the singleton instance of Config
func Get() *Config {
	configOnce.Do(func() {
		config = &Config{}
		if err := loadConfig(); err != nil {
			panic(fmt.Sprintf("Failed to load configuration: %v", err))
		}
	})
	return config
}

func loadConfig() error {
	viper.SetConfigFile(defaultPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("auth.username", "admin")
	viper.SetDefault("auth.password", "admin")
	viper.SetDefault("logging.environment", "development")
	viper.SetDefault("logging.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read configs file: %w", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal configs: %w", err)
	}

	return nil
}
