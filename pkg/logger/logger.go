package logger

import (
	"log/slog"
	"os"

	"github.com/histopathai/auth-service/pkg/config"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *config.LoggingConfig) *Logger {
	var handler slog.Handler

	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
	}
}

func parseLevel(level string) slog.Level {
	switch level {
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

// With returns a new Logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithGroup returns a new Logger with the given group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
	}
}
