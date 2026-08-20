package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/domain/usecase/process"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/retry"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	client    *Client
	processor *process.ImageProcessor
	log       *logger.Logger
	kit       observability.Kit
}

type messageHandler func(ctx context.Context, delivery amqp.Delivery) error

func NewConsumer(
	cfg config.RabbitMQConfig,
	imageProcessor *process.ImageProcessor,
	log *logger.Logger,
	kit observability.Kit,
) (*Consumer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		client:    client,
		processor: imageProcessor,
		log:       log,
		kit:       kit,
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
			err := consumer.kit.Run(msgCtx, "rabbitmq.consume", func(ctx context.Context) error {
				return handler(ctx, delivery)
			},
				observability.String("messaging.system", "rabbitmq"),
				observability.String("messaging.destination", queue),
				observability.String("messaging.message_id", delivery.MessageId),
			)
			if err != nil {
				status = "error"
				consumer.log.Error(msgCtx, "message processing failed", "queue", queue, "error", err)
				_ = delivery.Nack(false, false)
			} else {
				_ = delivery.Ack(false)
			}

			consumer.kit.Metrics().RecordRabbitMQConsumed(msgCtx, queue, status)
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
		consumer.kit.AddEvent(ctx, "retry",
			observability.Int("attempt", attempt),
			observability.String("error", err.Error()),
			observability.String("avatar_id", avatarID),
		)
		consumer.log.Warn(ctx, "processing failed", "avatar_id", avatarID, "attempt", attempt, "error", err)
	})
}

func (consumer *Consumer) Close() error {
	return consumer.client.Close()
}
