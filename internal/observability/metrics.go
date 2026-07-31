package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	metricsEnabled  bool
	durationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
)

type metricSet struct {
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
}

var metrics metricSet

func InitMetrics() error {
	err := registerMetrics()
	if err == nil {
		metricsEnabled = true
	}
	return err
}

func RecordHTTPRequest(method, path, status string, duration time.Duration) {
	if metrics.httpRequestsTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	)

	metrics.httpRequestsTotal.Add(context.Background(), 1, labels)
	metrics.httpRequestDuration.Record(context.Background(), duration.Seconds(), labels)
}

func RecordUpload(status string, duration time.Duration) {
	recordStatusCounterAndHistogram(metrics.uploadsTotal, metrics.uploadDuration, status, duration)
}

func RecordDelete(status string, duration time.Duration) {
	recordStatusCounterAndHistogram(metrics.deletesTotal, metrics.deleteDuration, status, duration)
}

func RecordProcessing(operation, status string, duration time.Duration) {
	if metrics.processingTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	)

	metrics.processingTotal.Add(context.Background(), 1, labels)
	metrics.processingDuration.Record(context.Background(), duration.Seconds(), labels)
}

func AddStorageBytes(bytes int64) {
	if metrics.storageUsageBytes != nil {
		metrics.storageUsageBytes.Add(context.Background(), float64(bytes))
	}
}

func SubStorageBytes(bytes int64) {
	if metrics.storageUsageBytes != nil {
		metrics.storageUsageBytes.Add(context.Background(), -float64(bytes))
	}
}

func RecordRabbitMQPublished(routingKey, status string) {
	recordLabeledCounter(metrics.rabbitMQMessagesPublished, "routing_key", routingKey, "status", status)
}

func RecordRabbitMQConsumed(queue, status string) {
	recordLabeledCounter(metrics.rabbitMQMessagesConsumed, "queue", queue, "status", status)
}

func RecordS3Operation(operation, status string, duration time.Duration) {
	if metrics.s3OperationsTotal == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", status),
	)

	metrics.s3OperationsTotal.Add(context.Background(), 1, labels)
	metrics.s3OperationDuration.Record(context.Background(), duration.Seconds(),
		metric.WithAttributes(attribute.String("operation", operation)))
}

func RecordDBQuery(operation string, duration time.Duration) {
	if metrics.dbQueryDuration == nil {
		return
	}

	metrics.dbQueryDuration.Record(context.Background(), duration.Seconds(),
		metric.WithAttributes(attribute.String("operation", operation)))
}

func recordStatusCounterAndHistogram(
	counter metric.Int64Counter,
	histogram metric.Float64Histogram,
	status string,
	duration time.Duration,
) {
	if counter == nil {
		return
	}

	labels := metric.WithAttributes(attribute.String("status", status))
	counter.Add(context.Background(), 1, labels)
	histogram.Record(context.Background(), duration.Seconds(), labels)
}

func recordLabeledCounter(
	counter metric.Int64Counter,
	firstKey, firstValue, secondKey, secondValue string,
) {
	if counter == nil {
		return
	}

	counter.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String(firstKey, firstValue),
		attribute.String(secondKey, secondValue),
	))
}

func registerMetrics() error {
	meter := otel.Meter(TracerName)
	var err error

	metrics.httpRequestsTotal, err = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		return err
	}

	metrics.httpRequestDuration, err = meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	metrics.uploadsTotal, err = meter.Int64Counter("avatars_uploads_total",
		metric.WithDescription("Total number of avatar uploads"))
	if err != nil {
		return err
	}

	metrics.uploadDuration, err = meter.Float64Histogram("avatars_upload_duration_seconds",
		metric.WithDescription("Avatar upload duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	metrics.deletesTotal, err = meter.Int64Counter("avatars_deletes_total",
		metric.WithDescription("Total number of avatar deletions"))
	if err != nil {
		return err
	}

	metrics.deleteDuration, err = meter.Float64Histogram("avatars_delete_duration_seconds",
		metric.WithDescription("Avatar delete duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	metrics.processingTotal, err = meter.Int64Counter("avatars_processing_total",
		metric.WithDescription("Total number of avatar processing operations"))
	if err != nil {
		return err
	}

	metrics.processingDuration, err = meter.Float64Histogram("avatars_processing_duration_seconds",
		metric.WithDescription("Avatar processing duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	metrics.storageUsageBytes, err = meter.Float64UpDownCounter("avatars_storage_bytes_total",
		metric.WithDescription("Total storage used by avatars across all users"))
	if err != nil {
		return err
	}

	metrics.rabbitMQMessagesPublished, err = meter.Int64Counter("rabbitmq_messages_published_total",
		metric.WithDescription("Total number of messages published to RabbitMQ"))
	if err != nil {
		return err
	}

	metrics.rabbitMQMessagesConsumed, err = meter.Int64Counter("rabbitmq_messages_consumed_total",
		metric.WithDescription("Total number of messages consumed from RabbitMQ"))
	if err != nil {
		return err
	}

	metrics.s3OperationsTotal, err = meter.Int64Counter("s3_operations_total",
		metric.WithDescription("Total number of S3 operations"))
	if err != nil {
		return err
	}

	metrics.s3OperationDuration, err = meter.Float64Histogram("s3_operation_duration_seconds",
		metric.WithDescription("S3 operation duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	metrics.dbQueryDuration, err = meter.Float64Histogram("db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return err
	}

	return nil
}
