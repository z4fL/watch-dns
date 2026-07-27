package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/z4fL/watch-dns/internal/config"
)

func New(cfg config.LogConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
