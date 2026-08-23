package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var durationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

type Metrics struct {
	meter                     metric.Meter
	httpRequestsTotal         metric.Int64Counter
	httpRequestDuration       metric.Float64Histogram
	uploadsTotal              metric.Int64Counter
	uploadDuration            metric.Float64Histogram
	deletesTotal              metric.Int64Counter
	deleteDuration            metric.Float64Histogram
	processingTotal           metric.Int64Counter
	processingDuration        metric.Float64Histogram
	storageUsageBytes         metric.Float64UpDownCounter
	rabbitMQMessagesPublished metric.Int64Counter
	rabbitMQMessagesConsumed  metric.Int64Counter
	s3OperationsTotal         metric.Int64Counter
	s3OperationDuration       metric.Float64Histogram
	dbQueryDuration           metric.Float64Histogram
	circuitBreakerState       metric.Int64Gauge
}

var (
	_ Recorder                 = (*Metrics)(nil)
	_ DatabaseMetricsRegistrar = (*Metrics)(nil)
)

func NewMetrics(meter metric.Meter) (*Metrics, error) {
	metrics := &Metrics{meter: meter}
	var err error

	metrics.httpRequestsTotal, err = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		return nil, err
	}

	metrics.httpRequestDuration, err = meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.uploadsTotal, err = meter.Int64Counter("avatars_uploads_total",
		metric.WithDescription("Total number of avatar uploads"))
	if err != nil {
		return nil, err
	}

	metrics.uploadDuration, err = meter.Float64Histogram("avatars_upload_duration_seconds",
		metric.WithDescription("Avatar upload duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.deletesTotal, err = meter.Int64Counter("avatars_deletes_total",
		metric.WithDescription("Total number of avatar deletions"))
	if err != nil {
		return nil, err
	}

	metrics.deleteDuration, err = meter.Float64Histogram("avatars_delete_duration_seconds",
		metric.WithDescription("Avatar delete duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.processingTotal, err = meter.Int64Counter("avatars_processing_total",
		metric.WithDescription("Total number of avatar processing operations"))
	if err != nil {
		return nil, err
	}

	metrics.processingDuration, err = meter.Float64Histogram("avatars_processing_duration_seconds",
		metric.WithDescription("Avatar processing duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.storageUsageBytes, err = meter.Float64UpDownCounter("avatars_storage_bytes_total",
		metric.WithDescription("Total storage used by avatars across all users"))
	if err != nil {
		return nil, err
	}

	metrics.rabbitMQMessagesPublished, err = meter.Int64Counter("rabbitmq_messages_published_total",
		metric.WithDescription("Total number of messages published to RabbitMQ"))
	if err != nil {
		return nil, err
	}

	metrics.rabbitMQMessagesConsumed, err = meter.Int64Counter("rabbitmq_messages_consumed_total",
		metric.WithDescription("Total number of messages consumed from RabbitMQ"))
	if err != nil {
		return nil, err
	}

	metrics.s3OperationsTotal, err = meter.Int64Counter("s3_operations_total",
		metric.WithDescription("Total number of S3 operations"))
	if err != nil {
		return nil, err
	}

	metrics.s3OperationDuration, err = meter.Float64Histogram("s3_operation_duration_seconds",
		metric.WithDescription("S3 operation duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.dbQueryDuration, err = meter.Float64Histogram("db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, err
	}

	metrics.circuitBreakerState, err = meter.Int64Gauge("circuit_breaker_state",
		metric.WithDescription("Circuit breaker state: 0 closed, 1 half-open, 2 open"))
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (metrics *Metrics) RecordHTTPRequest(ctx context.Context, method, path, status string, duration time.Duration) {
	if metrics == nil || metrics.httpRequestsTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	)

	metrics.httpRequestsTotal.Add(ctx, 1, labels)
	metrics.httpRequestDuration.Record(ctx, duration.Seconds(), labels)
}

func (metrics *Metrics) RecordUpload(ctx context.Context, status string, duration time.Duration) {
	metrics.recordStatusCounterAndHistogram(ctx, metrics.uploadsTotal, metrics.uploadDuration, status, duration)
}

func (metrics *Metrics) RecordDelete(ctx context.Context, status string, duration time.Duration) {
	metrics.recordStatusCounterAndHistogram(ctx, metrics.deletesTotal, metrics.deleteDuration, status, duration)
}

func (metrics *Metrics) RecordProcessing(ctx context.Context, operation, status string, duration time.Duration) {
	if metrics == nil || metrics.processingTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	)

	metrics.processingTotal.Add(ctx, 1, labels)
	metrics.processingDuration.Record(ctx, duration.Seconds(), labels)
}

func (metrics *Metrics) AddStorageBytes(ctx context.Context, bytes int64) {
	if metrics != nil && metrics.storageUsageBytes != nil {
		metrics.storageUsageBytes.Add(ctx, float64(bytes))
	}
}

func (metrics *Metrics) SubStorageBytes(ctx context.Context, bytes int64) {
	if metrics != nil && metrics.storageUsageBytes != nil {
		metrics.storageUsageBytes.Add(ctx, -float64(bytes))
	}
}

func (metrics *Metrics) RecordRabbitMQPublished(ctx context.Context, routingKey, status string) {
	metrics.recordLabeledCounter(ctx, metrics.rabbitMQMessagesPublished, "routing_key", routingKey, "status", status)
}

func (metrics *Metrics) RecordRabbitMQConsumed(ctx context.Context, queue, status string) {
	metrics.recordLabeledCounter(ctx, metrics.rabbitMQMessagesConsumed, "queue", queue, "status", status)
}

func (metrics *Metrics) RecordS3Operation(ctx context.Context, operation, status string, duration time.Duration) {
	if metrics == nil || metrics.s3OperationsTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	)

	metrics.s3OperationsTotal.Add(ctx, 1, labels)
	metrics.s3OperationDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("operation", operation)))
}

func (metrics *Metrics) RecordDBQuery(ctx context.Context, operation string, duration time.Duration) {
	if metrics == nil || metrics.dbQueryDuration == nil {
		return
	}

	metrics.dbQueryDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("operation", operation)))
}

func (metrics *Metrics) RecordCircuitBreakerState(ctx context.Context, name string, state int64) {
	if metrics == nil || metrics.circuitBreakerState == nil {
		return
	}

	metrics.circuitBreakerState.Record(ctx, state,
		metric.WithAttributes(attribute.String("name", name)))
}

func (metrics *Metrics) recordStatusCounterAndHistogram(
	ctx context.Context,
	counter metric.Int64Counter,
	histogram metric.Float64Histogram,
	status string,
	duration time.Duration,
) {
	if metrics == nil || counter == nil {
		return
	}

	labels := metric.WithAttributes(attribute.String("status", status))
	counter.Add(ctx, 1, labels)
	histogram.Record(ctx, duration.Seconds(), labels)
}

func (metrics *Metrics) recordLabeledCounter(
	ctx context.Context,
	counter metric.Int64Counter,
	firstKey, firstValue, secondKey, secondValue string,
) {
	if metrics == nil || counter == nil {
		return
	}

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String(firstKey, firstValue),
		attribute.String(secondKey, secondValue),
	))
}
