package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once sync.Once
)

type Config struct {
	Environment string
	LogLevel    string
}

func Initialize(evironment, logLevel string) {
	once.Do(func() {
		var handler slog.Handler
		opts := &slog.HandlerOptions{
			AddSource: true,
			Level:     getLogLevel(logLevel),
		}

		if evironment == "production" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = NewPrettyHandler(getLogLevel(logLevel))
		}

		logger := slog.New(handler)
		slog.SetDefault(logger) // Set as default logger for slog package
	})
}

func getLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
