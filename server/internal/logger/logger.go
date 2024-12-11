package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"
)

type mode int

const (
	ModeDefault mode = iota
	ModeDebug
)

var keys = []string{"TraceID"}

type MyLogHandler struct {
	slog.Handler
}

var _ slog.Handler = &MyLogHandler{}

func (h *MyLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		for _, key := range keys {
			if v := ctx.Value(key); v != nil {
				r.AddAttrs(slog.Attr{Key: string(key), Value: slog.AnyValue(v)})
			}
		}
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogger(mode mode, output io.Writer) *slog.Logger {
	var logLevel slog.Level
	switch mode {
	case ModeDebug:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(&MyLogHandler{slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	})})
}

func Info(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	logf(ctx, slog.LevelInfo, format, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	logf(ctx, slog.LevelDebug, format, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	logf(ctx, slog.LevelError, format, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	logf(ctx, slog.LevelWarn, format, args...)
}

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logCommon(ctx, level, msg, false, args...)
}

func logf(ctx context.Context, level slog.Level, format string, args ...any) {
	logCommon(ctx, level, format, true, args...)
}

func logCommon(ctx context.Context, level slog.Level, msg string, isFormat bool, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// skip 3 frames [Callers, Info|Debug|Error|Warn, logCommon]
	runtime.Callers(3, pcs[:])

	var r slog.Record
	if isFormat {
		r = slog.NewRecord(time.Now(), level, fmt.Sprintf(msg, args...), pcs[0])
	} else {
		r = slog.NewRecord(time.Now(), level, msg, pcs[0])
		r.Add(args...)
	}

	_ = logger.Handler().Handle(ctx, r)
}
