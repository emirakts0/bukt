package logger

import (
	"github.com/phsym/console-slog"
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

func Initialize(cfg Config) {
	once.Do(func() {
		var handler slog.Handler

		//Dynamic
		logLevel := &slog.LevelVar{}
		logLevel.Set(getLogLevel(cfg.LogLevel))

		if cfg.Environment == "production" {
			handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: false,
				Level:     logLevel,
			})
		} else {
			handler = console.NewHandler(os.Stdout, &console.HandlerOptions{
				Level:     logLevel,
				AddSource: true,
			})
		}

		logger := slog.New(handler)
		slog.SetDefault(logger)
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
