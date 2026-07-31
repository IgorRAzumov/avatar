package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/processor"
	"avatar/internal/retry"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

type Consumer struct {
	client    *Client
	processor *processor.ImageProcessor
	log       *logger.Logger
}

type messageHandler func(ctx context.Context, delivery amqp.Delivery) error

func NewConsumer(cfg config.RabbitMQConfig, imageProcessor *processor.ImageProcessor, log *logger.Logger) (*Consumer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		client:    client,
		processor: imageProcessor,
		log:       log,
	}, nil
}

func (consumer *Consumer) Run(ctx context.Context) error {
	uploadErr := make(chan error, 1)
	deleteErr := make(chan error, 1)

	go func() {
		uploadErr <- consumer.consume(ctx, config.DefaultRabbitMQUploadQueue, consumer.handleUpload)
	}()
	go func() {
		deleteErr <- consumer.consume(ctx, config.DefaultRabbitMQDeleteQueue, consumer.handleDelete)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-uploadErr:
		return err
	case err := <-deleteErr:
		return err
	}
}

func (consumer *Consumer) consume(ctx context.Context, queue string, handler messageHandler) error {
	deliveries, err := consumer.client.channel.Consume(
		queue,
		fmt.Sprintf("avatar-worker-%s", queue),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", queue, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed for queue %q", queue)
			}
			status := "success"
			msgCtx := extractTraceContext(ctx, delivery.Headers)
			err := observability.Run(msgCtx, "rabbitmq.consume", func(ctx context.Context) error {
				return handler(ctx, delivery)
			},
				attribute.String("messaging.system", "rabbitmq"),
				attribute.String("messaging.destination", queue),
				attribute.String("messaging.message_id", delivery.MessageId),
			)
			if err != nil {
				status = "error"
				consumer.log.WithContext(msgCtx).Error("message processing failed", "queue", queue, "error", err)
				_ = delivery.Nack(false, false)
			} else {
				_ = delivery.Ack(false)
			}

			observability.RecordRabbitMQConsumed(queue, status)
		}
	}
}

func (consumer *Consumer) handleUpload(ctx context.Context, delivery amqp.Delivery) error {
	var event model.AvatarUploadEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return fmt.Errorf("decode upload event: %w", err)
	}
	return consumer.process(ctx, event.AvatarID, func() error {
		return consumer.processor.ProcessUpload(ctx, event)
	})
}

func (consumer *Consumer) handleDelete(ctx context.Context, delivery amqp.Delivery) error {
	var event model.AvatarDeleteEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return fmt.Errorf("decode delete event: %w", err)
	}
	return consumer.process(ctx, event.AvatarID, func() error {
		return consumer.processor.ProcessDelete(ctx, event)
	})
}

func (consumer *Consumer) process(ctx context.Context, avatarID string, fn func() error) error {
	return retry.WithBackoff(ctx, retry.DefaultMaxAttempts, fn, func(attempt int, err error) {
		observability.AddEvent(ctx, "retry",
			attribute.Int("attempt", attempt),
			attribute.String("error", err.Error()),
			attribute.String("avatar_id", avatarID),
		)
		consumer.log.WithContext(ctx).Warn("processing failed", "avatar_id", avatarID, "attempt", attempt, "error", err)
	})
}

func (consumer *Consumer) Close() error {
	return consumer.client.Close()
}
