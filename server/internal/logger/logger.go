package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

const (
	MaxDepth = 50
	Skip     = 5
)

var (
	keys = []string{}
)

type mode int

const (
	ModeDefault mode = iota
	ModeDebug
)

type MyLogHandler struct {
	slog.Handler
	hasStack bool
}

var _ slog.Handler = &MyLogHandler{}

func AddKey(newkeys ...string) {
	keys = append(keys, newkeys...)
}

func (h *MyLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.hasStack {
		stackInfo := slog.Any("stackTrace", stackTrace(getStack()))
		r.Add(stackInfo)
	}

	if ctx != nil {
		for _, key := range keys {
			if v := ctx.Value(key); v != nil {
				r.AddAttrs(slog.Attr{Key: string(key), Value: slog.AnyValue(v)})
			}
		}
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogger(mode mode, hasStack bool, output io.Writer) *slog.Logger {
	var logLevel slog.Level
	switch mode {
	case ModeDebug:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(&MyLogHandler{
		Handler: slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: false,
		}),
		hasStack: hasStack})
}

func SetupLogger(mode mode, hasStack bool, output io.Writer) {
	customLogger := NewLogger(mode, hasStack, output)
	slog.SetDefault(customLogger)
}

func I(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, nil, args...)
}

func D(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, nil, args...)
}

func E(ctx context.Context, msg string, err error, args ...any) {
	log(ctx, slog.LevelError, msg, err, args...)
}

func W(ctx context.Context, msg string, err error, args ...any) {
	log(ctx, slog.LevelWarn, msg, err, args...)
}

func log(ctx context.Context, level slog.Level, msg string, err error, args ...any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}

	stack := make([]uintptr, MaxDepth)

	var r slog.Record
	r = slog.NewRecord(time.Now(), level, msg, stack[0])

	if err != nil {
		var errInfo slog.Attr
		// error info
		errInfo = slog.Any("error", parseErrorString(err.Error()))
		args = append(args, errInfo)
	}

	r.Add(args...)

	_ = logger.Handler().Handle(ctx, r)
}

func getStack() []uintptr {
	stack := make([]uintptr, 50)
	length := runtime.Callers(Skip, stack)
	return stack[:length]
}

func stackTrace(stack []uintptr) []map[string]interface{} {
	frames := runtime.CallersFrames(stack)
	res := make([]map[string]interface{}, 0)
	for {
		frame, more := frames.Next()
		res = append(res, map[string]interface{}{
			"file":     frame.File,
			"line":     frame.Line,
			"function": frame.Function,
		})

		if !more {
			break
		}
	}
	return res
}
func parseErrorString(input string) []slog.Attr {
	parts := strings.Split(input, ", ")
	// ここが要検討
	// エラーメッセージが1つの場合は、エラーメッセージをそのまま返す
	if len(parts) == 1 {
		return []slog.Attr{slog.String("message", input)}
	}

	result := make([]slog.Attr, 0, len(parts))

	for _, part := range parts {
		kv := strings.SplitN(part, ": ", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.Trim(kv[1], `"`)
		if key == "code" {
			var code int
			fmt.Sscanf(value, "%d", &code)
			result = append(result, slog.Int(key, code))
		} else {
			result = append(result, slog.String(key, value))
		}
	}

	return result
}
