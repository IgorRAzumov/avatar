package rabbitmq_test

import (
	"context"
	"testing"

	"avatar/internal/adapter/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestInjectExtractTraceContextRoundTrip(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	ctx, span := otel.Tracer("test").Start(context.Background(), "publish")
	defer span.End()

	headers := amqp.Table{}
	rabbitmq.InjectTraceContextForTest(ctx, headers)
	require.NotEmpty(t, headers)

	extracted := rabbitmq.ExtractTraceContextForTest(context.Background(), headers)
	extractedSpan := trace.SpanFromContext(extracted)
	require.True(t, extractedSpan.SpanContext().IsValid())
	require.Equal(t, span.SpanContext().TraceID(), extractedSpan.SpanContext().TraceID())
	require.Equal(t, span.SpanContext().SpanID(), extractedSpan.SpanContext().SpanID())
}

func TestExtractTraceContextEmptyHeaders(t *testing.T) {
	ctx := rabbitmq.ExtractTraceContextForTest(context.Background(), nil)
	require.False(t, trace.SpanFromContext(ctx).SpanContext().IsValid())
}
