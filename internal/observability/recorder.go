package observability

import (
	"context"
	"time"
)

type Recorder interface {
	RecordHTTPRequest(ctx context.Context, method, path, status string, duration time.Duration)
	RecordUpload(ctx context.Context, status string, duration time.Duration)
	RecordDelete(ctx context.Context, status string, duration time.Duration)
	RecordProcessing(ctx context.Context, operation, status string, duration time.Duration)
	AddStorageBytes(ctx context.Context, bytes int64)
	SubStorageBytes(ctx context.Context, bytes int64)
	RecordRabbitMQPublished(ctx context.Context, routingKey, status string)
	RecordRabbitMQConsumed(ctx context.Context, queue, status string)
	RecordS3Operation(ctx context.Context, operation, status string, duration time.Duration)
	RecordDBQuery(ctx context.Context, operation string, duration time.Duration)
	RecordCircuitBreakerState(ctx context.Context, name string, state int64)
}

type DatabaseMetricsRegistrar interface {
	RegisterDatabaseMetrics(stats DatabaseStats) (unregister func(), err error)
}

type noopRecorder struct{}

var NopRecorder Recorder = noopRecorder{}

func (noopRecorder) RecordHTTPRequest(context.Context, string, string, string, time.Duration) {}
func (noopRecorder) RecordUpload(context.Context, string, time.Duration)                      {}
func (noopRecorder) RecordDelete(context.Context, string, time.Duration)                      {}
func (noopRecorder) RecordProcessing(context.Context, string, string, time.Duration)          {}
func (noopRecorder) AddStorageBytes(context.Context, int64)                                   {}
func (noopRecorder) SubStorageBytes(context.Context, int64)                                   {}
func (noopRecorder) RecordRabbitMQPublished(context.Context, string, string)                  {}
func (noopRecorder) RecordRabbitMQConsumed(context.Context, string, string)                   {}
func (noopRecorder) RecordS3Operation(context.Context, string, string, time.Duration)         {}
func (noopRecorder) RecordDBQuery(context.Context, string, time.Duration)                     {}
func (noopRecorder) RecordCircuitBreakerState(context.Context, string, int64)                 {}

func (noopRecorder) RegisterDatabaseMetrics(DatabaseStats) (func(), error) {
	return func() {}, nil
}
