package logger

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	log *slog.Logger
}

type Option func(*loggerOptions)

type loggerOptions struct {
	serviceName string
	logProvider *log.LoggerProvider
}

// WithOTLP enables OpenTelemetry log export via the configured LoggerProvider.
func WithOTLP(provider *log.LoggerProvider, serviceName string) Option {
	return func(opts *loggerOptions) {
		opts.logProvider = provider
		opts.serviceName = serviceName
	}
}

func New(out io.Writer, level slog.Level, options ...Option) *Logger {
	cfg := loggerOptions{}
	for _, option := range options {
		option(&cfg)
	}

	handlers := []slog.Handler{slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})}
	if cfg.logProvider != nil {
		handlers = append(handlers, otelslog.NewHandler(cfg.serviceName, otelslog.WithLoggerProvider(cfg.logProvider)))
	}

	return &Logger{log: slog.New(newMultiHandler(handlers...))}
}

func Nop() *Logger {
	return &Logger{log: slog.New(slog.DiscardHandler)}
}

func (logger *Logger) Info(ctx context.Context, msg string, args ...any) {
	logger.logAt(ctx, slog.LevelInfo, msg, args...)
}

func (logger *Logger) Warn(ctx context.Context, msg string, args ...any) {
	logger.logAt(ctx, slog.LevelWarn, msg, args...)
}

func (logger *Logger) Error(ctx context.Context, msg string, args ...any) {
	logger.logAt(ctx, slog.LevelError, msg, args...)
}

func (logger *Logger) Debug(ctx context.Context, msg string, args ...any) {
	logger.logAt(ctx, slog.LevelDebug, msg, args...)
}

func (logger *Logger) logAt(ctx context.Context, level slog.Level, msg string, args ...any) {
	log := logger.log
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		log = log.With(
			"trace_id", spanCtx.TraceID().String(),
			"span_id", spanCtx.SpanID().String(),
		)
	}
	log.Log(ctx, level, msg, args...)
}

// WithService returns a logger with a static service attribute.
func (logger *Logger) WithService(service string) *Logger {
	return &Logger{log: logger.log.With("service", service)}
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	filtered := make([]slog.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if handler != nil {
			filtered = append(filtered, handler)
		}
	}
	return multiHandler{handlers: filtered}
}

func (handler multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, item := range handler.handlers {
		if item.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (handler multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, item := range handler.handlers {
		if err := item.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (handler multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(handler.handlers))
	for index, item := range handler.handlers {
		next[index] = item.WithAttrs(attrs)
	}
	return multiHandler{handlers: next}
}

func (handler multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(handler.handlers))
	for index, item := range handler.handlers {
		next[index] = item.WithGroup(name)
	}
	return multiHandler{handlers: next}
}
