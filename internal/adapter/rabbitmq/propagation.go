package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type amqpHeadersCarrier amqp.Table

func (carrier amqpHeadersCarrier) Get(key string) string {
	value, ok := carrier[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func (carrier amqpHeadersCarrier) Set(key, value string) {
	carrier[key] = value
}

func (carrier amqpHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(carrier))
	for key := range carrier {
		keys = append(keys, key)
	}
	return keys
}

func injectTraceContext(ctx context.Context, headers amqp.Table) {
	if headers == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, amqpHeadersCarrier(headers))
}

func extractTraceContext(ctx context.Context, headers amqp.Table) context.Context {
	if len(headers) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, amqpHeadersCarrier(headers))
}
