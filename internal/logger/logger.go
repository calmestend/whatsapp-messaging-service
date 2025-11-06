package logger

import (
	"io"
	"log/slog"
	"os"
)

func Init(filename string) *os.File {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	w := io.MultiWriter(os.Stderr, file)
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	return file
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

func LogResponse(status int, uri string, body map[string]any) {
	attr := []slog.Attr{
		slog.Int("status", status),
		slog.String("uri", uri),
		slog.Any("body", body),
	}

	groupAttrs := make([]any, len(attr))
	for i, a := range attr {
		groupAttrs[i] = a
	}

	switch {
	case status >= 500:
		Error("HTTP Response", groupAttrs...)
	case status >= 400:
		Warn("HTTP Response", groupAttrs...)
	default:
		Info("HTTP Response", groupAttrs...)
	}
}
