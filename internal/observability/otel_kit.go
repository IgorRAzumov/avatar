package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type otelKit struct {
	tracer  trace.Tracer
	metrics Recorder
}

func NewOTelKit(tracer trace.Tracer, metrics Recorder) Kit {
	if metrics == nil {
		metrics = NopRecorder
	}
	return &otelKit{tracer: tracer, metrics: metrics}
}

func NewTestKit() Kit {
	return NewOTelKit(sdktrace.NewTracerProvider().Tracer(TracerName), NopRecorder)
}

func (kit *otelKit) Metrics() Recorder {
	return kit.metrics
}

func (kit *otelKit) Run(
	ctx context.Context,
	name string,
	fn func(context.Context) error,
	attrs ...Attr,
) error {
	ctx, span := kit.tracer.Start(ctx, name, trace.WithAttributes(attrsToOTel(attrs)...))
	defer span.End()

	err := fn(ctx)
	if err != nil {
		kit.recordError(span, err)
	}
	return err
}

func (kit *otelKit) RunDB(
	ctx context.Context,
	operation string,
	fn func(context.Context) error,
	attrs ...Attr,
) error {
	return kit.Run(ctx, operation, func(ctx context.Context) error {
		start := time.Now()
		err := fn(ctx)
		kit.metrics.RecordDBQuery(ctx, operation, time.Since(start))
		return err
	}, attrs...)
}

func (kit *otelKit) RunS3(
	ctx context.Context,
	spanName, operation string,
	fn func(context.Context) error,
	attrs ...Attr,
) error {
	return kit.Run(ctx, spanName, func(ctx context.Context) error {
		start := time.Now()
		fnErr := fn(ctx)
		status := "success"
		if fnErr != nil {
			status = "error"
		}
		kit.metrics.RecordS3Operation(ctx, operation, status, time.Since(start))
		return fnErr
	}, attrs...)
}

func (kit *otelKit) RegisterDatabaseMetrics(stats DatabaseStats) (func(), error) {
	if registrar, ok := kit.metrics.(DatabaseMetricsRegistrar); ok {
		return registrar.RegisterDatabaseMetrics(stats)
	}
	return func() {}, nil
}

func (kit *otelKit) SetAttributes(ctx context.Context, attrs ...Attr) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.SetAttributes(attrsToOTel(attrs)...)
}

func (kit *otelKit) AddEvent(ctx context.Context, name string, attrs ...Attr) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrsToOTel(attrs)...))
}

func (kit *otelKit) recordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
