package logger

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env string) *Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		panic("invalid environment")
	}

	return &Logger{logger: logger}
}

type Logger struct {
	logger *slog.Logger
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args)
}
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args)
}
