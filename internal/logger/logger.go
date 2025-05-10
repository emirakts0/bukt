package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	logger      *zap.Logger
	sugared     *zap.SugaredLogger
	atomicLevel = zap.NewAtomicLevel()
	once        sync.Once
)

type Config struct {
	Environment string
	LogLevel    string
}

// Initialize sets up the logger with the given configuration
func Initialize(cfg Config) {
	once.Do(func() {
		zapCfg := zap.Config{
			Level:       atomicLevel,
			Development: cfg.Environment != "production",
			Encoding:    getEncoding(cfg.Environment),
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "timestamp",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		if cfg.Environment == "production" {
			zapCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		}

		atomicLevel.SetLevel(getZapLevel(cfg.LogLevel))

		var err error
		logger, err = zapCfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}

		sugared = logger.Sugar()
	})
}

// Get returns the singleton logger instance
func Get() *zap.Logger {
	if logger == nil {
		Initialize(Config{
			Environment: "development",
			LogLevel:    "info",
		})
	}
	return logger
}

// GetSugared returns the singleton SugaredLogger instance
func GetSugared() *zap.SugaredLogger {
	if sugared == nil {
		Initialize(Config{
			Environment: "development",
			LogLevel:    "info",
		})
	}
	return sugared
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// getZapLevel maps string log level to zapcore.Level
func getZapLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEncoding determines the encoding based on environment
func getEncoding(env string) string {
	if env == "production" {
		return "json"
	}
	return "console"
}
