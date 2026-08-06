package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"avatar/internal/logger"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestLoggerWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, slog.LevelInfo)

	log.Info(context.Background(), "hello", "key", "value")

	out := buf.String()
	assert.True(t, strings.Contains(out, "hello"))
	assert.True(t, strings.Contains(out, "key"))
	assert.True(t, strings.Contains(out, "value"))
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, slog.LevelDebug)

	log.Info(context.Background(), "i")
	log.Warn(context.Background(), "w")
	log.Error(context.Background(), "e")

	out := buf.String()
	assert.Contains(t, out, "\"i\"")
	assert.Contains(t, out, "\"w\"")
	assert.Contains(t, out, "\"e\"")
}

func TestNopLoggerDoesNotPanic(t *testing.T) {
	log := logger.Nop()
	assert.NotPanics(t, func() {
		log.Info(context.Background(), "i")
		log.Warn(context.Background(), "w")
		log.Error(context.Background(), "e")
	})
}

func TestContextAddsTraceFields(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	var buf bytes.Buffer
	log := logger.New(&buf, slog.LevelInfo)

	ctx, span := otel.Tracer("test").Start(context.Background(), "test-span")
	defer span.End()

	log.Info(ctx, "correlated")

	out := buf.String()
	assert.Contains(t, out, "trace_id")
	assert.Contains(t, out, "span_id")
}

func TestContextAddsRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, slog.LevelInfo)

	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "req-123")
	log.Info(ctx, "request")

	out := buf.String()
	assert.Contains(t, out, "request_id")
	assert.Contains(t, out, "req-123")
}
