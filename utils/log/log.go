package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

type DepthHandler struct {
	slog.Handler
	depth int
}

func NewDepthHandler(inner slog.Handler, depth int) slog.Handler {
	return &DepthHandler{
		Handler: inner,
		depth:   depth,
	}
}

func (h *DepthHandler) Handle(ctx context.Context, r slog.Record) error {
	if _, file, line, ok := runtime.Caller(h.depth); ok {
		source := fmt.Sprintf("%s:%d", filepath.Base(file), line)
		r.Add("source", source)
	}
	return h.Handler.Handle(ctx, r)
}

func LogSet() {
	base := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false, //소스코드 파일 출력
		Level:     slog.LevelInfo,
	})
	handler := NewDepthHandler(base, 4)

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func Fatal(format string, args ...any) {
	slog.Error(format, args...)
	os.Exit(1)
}

func Info(format string, args ...any) {
	slog.Info(format, args...)
}

func Debug(format string, args ...any) {
	slog.Debug(format, args...)
}

func Warn(format string, args ...any) {
	slog.Warn(format, args...)
}

func Error(format string, args ...any) {
	slog.Error(format, args...)
}
