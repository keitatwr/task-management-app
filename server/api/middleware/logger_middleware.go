package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoggerConfig struct {
	BaseLogLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level
}

type Option func(*LoggerConfig)

func NewLoggerConfig(opts ...Option) *LoggerConfig {
	lc := &LoggerConfig{}
	for _, opt := range opts {
		opt(lc)
	}
	return lc
}

func WithBaseLogLevel(level slog.Level) Option {
	return func(c *LoggerConfig) {
		c.BaseLogLevel = level
	}
}

func WithClientErrorLogLevel(level slog.Level) Option {
	return func(c *LoggerConfig) {
		c.ClientErrorLevel = level
	}
}

func WithServerErrorLogLevel(level slog.Level) Option {
	return func(c *LoggerConfig) {
		c.ServerErrorLevel = level
	}
}

var (
	requestIDHeaderKey = "X-Request-ID"
	timeFormater       = "2006/1/2 15:04:05.000 JTS"
)

func LoggingMiddleware(config *LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			slog.Error("Failed to load location for Asia/Tokyo", "error", err)
		}

		start := time.Now().In(jst)
		startStr := start.Format(timeFormater)
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		traceID, err := uuid.NewRandom()
		if err != nil {
			slog.Error("Failed to generate trace ID", "error", err)
		}
		traceIDStr := traceID.String()

		ctx := context.WithValue(c.Request.Context(), "TraceID", traceIDStr)
		c.Request = c.Request.WithContext(ctx)

		params := map[string]string{}
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}

		logRequest(c, config, traceIDStr, startStr, method, path, query, clientIP, userAgent, params)

		c.Next()

		end := time.Now().In(jst)
		endStr := end.Format(timeFormater)
		latency := end.Sub(start)
		status := c.Writer.Status()

		logResponse(c, config, traceIDStr, endStr, latency, status)
	}
}

func logRequest(c *gin.Context, config *LoggerConfig,
	traceIDStr, startStr, method, path, query, clientIP, userAgent string,
	params map[string]string) {
	requestAttributes := []slog.Attr{
		slog.String("Time", startStr),
		slog.String("TraceID", traceIDStr),
		slog.String("Method", method),
		slog.String("Path", path),
		slog.String("Query", query),
		slog.Any("Params", params),
		slog.String("ClientIP", clientIP),
		slog.String("UserAgent", userAgent),
	}

	loggerWithTraceID := slog.With("TraceID", traceIDStr)

	loggerWithTraceID.LogAttrs(
		c.Request.Context(),
		config.BaseLogLevel,
		"Request Log",
		requestAttributes...,
	)
}

func logResponse(c *gin.Context, config *LoggerConfig,
	traceIDStr, endStr string, latency time.Duration, status int) {
	logLevel := determineLogLevel(status, config)

	responseAttributes := []slog.Attr{
		slog.String("Time", endStr),
		slog.String("Latency", convertLatency(latency)),
		slog.Int("Status", status),
	}

	loggerWithTraceID := slog.With("TraceID", traceIDStr)

	loggerWithTraceID.LogAttrs(
		c.Request.Context(),
		logLevel,
		"Response Log",
		responseAttributes...,
	)
}

func determineLogLevel(status int, config *LoggerConfig) slog.Level {
	if status >= 400 && status < 500 {
		return config.ClientErrorLevel
	} else if status >= 500 {
		return config.ServerErrorLevel
	}
	return config.BaseLogLevel
}

func convertLatency(latency time.Duration) string {
	return latency.String()
}
