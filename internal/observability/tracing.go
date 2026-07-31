package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Tracer() trace.Tracer {
	return otel.Tracer(TracerName)
}

// Run executes fn inside a span and records errors on failure.
func Run(
	ctx context.Context,
	name string,
	fn func(context.Context) error,
	attrs ...attribute.KeyValue,
) error {
	ctx, span := Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
	defer span.End()

	err := fn(ctx)
	if err != nil {
		RecordError(span, err)
	}
	return err
}

// RunResult executes fn inside a span and returns its result.
func RunResult[T any](
	ctx context.Context,
	name string,
	fn func(context.Context) (T, error),
	attrs ...attribute.KeyValue,
) (T, error) {
	ctx, span := Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		RecordError(span, err)
	}
	return result, err
}

func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.SetAttributes(attrs...)
}

func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
