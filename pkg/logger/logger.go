package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func init() {
	SetDebugLogger()
}

func SetDebugLogger() {
	timeFormat := "[" + time.RFC1123 + "]"
	handler := tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug, TimeFormat: timeFormat})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}
