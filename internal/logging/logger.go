package logging

import (
	"log/slog"
	"os"

	"stock-lens/internal/config"
)

func NewLogger(cfg config.LogConfig) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(parseLevel(cfg.Level))

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
