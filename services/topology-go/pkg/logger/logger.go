package logger

import (
	"log/slog"
	"os"
)

// New returns a structured JSON logger configured for production.
// In debug mode (LOG_LEVEL=debug), it emits human-readable text instead.
func New() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if os.Getenv("GIN_MODE") == "release" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
