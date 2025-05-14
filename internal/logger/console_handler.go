package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Gray   = "\033[90m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Pink   = "\033[36m"
)

// PrettyHandler only for development purposes, it will print logs in a pretty format.
type PrettyHandler struct {
	minLevel slog.Level
}

func NewPrettyHandler(level slog.Level) *PrettyHandler {
	return &PrettyHandler{minLevel: level}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format(time.RFC3339)

	var source string
	if r.PC != 0 {
		if fn := runtime.FuncForPC(r.PC); fn != nil {
			_, line := fn.FileLine(r.PC)
			source = fmt.Sprintf("%s:%d", fn.Name(), line)
		}
	}

	coloredTime := Purple + ts + Reset
	coloredSource := Blue + source + Reset

	coloredLevel := colorizeLevel(r.Level.String())

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	logLine := fmt.Sprintf("[%s] [%s] %s | %s [%s]\n",
		coloredTime,
		coloredLevel,
		r.Message,
		strings.Join(attrs, ", "),
		coloredSource,
	)

	_, err := os.Stdout.Write([]byte(logLine))
	return err
}

func colorizeLevel(level string) string {
	switch level {
	case "DEBUG":
		return Gray + "DEBUG" + Reset
	case "INFO":
		return Green + " INFO" + Reset
	case "WARN":
		return Yellow + " WARN" + Reset
	case "ERROR":
		return Red + "ERROR" + Reset
	default:
		return level
	}
}

func (h *PrettyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(_ string) slog.Handler {
	return h
}
