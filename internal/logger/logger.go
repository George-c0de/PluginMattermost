// Package logger Базовая структура для работы с логером
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

// SetupLogger инициализирует и возвращает указатель на Logger в зависимости от переданного окружения
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

// Logger Стурктура для работы с Logger
type Logger struct {
	logger *slog.Logger
}

// Info Запись уровня Info
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Error Запись уровня Error
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Debug Запись уровня Debug
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}
