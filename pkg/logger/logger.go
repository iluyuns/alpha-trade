package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Info logs a message at Info level.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Infof logs a formatted message at Info level.
func Infof(format string, args ...any) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

// Error logs a message at Error level.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Errorf logs a formatted message at Error level.
func Errorf(format string, args ...any) {
	defaultLogger.Error(fmt.Sprintf(format, args...))
}

// Debug logs a message at Debug level.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Debugf logs a formatted message at Debug level.
func Debugf(format string, args ...any) {
	defaultLogger.Debug(fmt.Sprintf(format, args...))
}

// Warn logs a message at Warn level.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Warnf logs a formatted message at Warn level.
func Warnf(format string, args ...any) {
	defaultLogger.Warn(fmt.Sprintf(format, args...))
}

// SetLevel sets the global logging level.
func SetLevel(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}
