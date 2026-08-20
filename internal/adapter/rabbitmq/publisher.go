package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"avatar/internal/circuitbreaker"
	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/observability"

	amqp "github.com/rabbitmq/amqp091-go"
)

const breakerName = "rabbitmq"

type Publisher struct {
	client  *Client
	kit     observability.Kit
	breaker *circuitbreaker.Breaker
}

func NewPublisher(cfg config.RabbitMQConfig, kit observability.Kit) (*Publisher, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		client: client,
		kit:    kit,
		breaker: circuitbreaker.New(circuitbreaker.Options{
			Name:          breakerName,
			OnStateChange: circuitbreaker.ReportStateTo(kit.Metrics()),
		}),
	}, nil
}

func (publisher *Publisher) PublishUploadEvent(ctx context.Context, event model.AvatarUploadEvent) error {
	return publisher.publish(ctx, config.DefaultRabbitMQUploadRoutingKey, event.AvatarID, event)
}

func (publisher *Publisher) PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error {
	return publisher.publish(ctx, config.DefaultRabbitMQDeleteRoutingKey, event.AvatarID, event)
}

func (publisher *Publisher) publish(ctx context.Context, routingKey, messageID string, payload any) error {
	status := "success"

	err := publisher.breaker.Do(func() error {
		return publisher.kit.Run(ctx, "rabbitmq.publish", func(ctx context.Context) error {
			body, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal event: %w", err)
			}

			headers := amqp.Table{}
			injectTraceContext(ctx, headers)

			return publisher.client.channel.PublishWithContext(
				ctx,
				publisher.client.exchange,
				routingKey,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/json",
					DeliveryMode: amqp.Persistent,
					MessageId:    messageID,
					Body:         body,
					Headers:      headers,
				},
			)
		},
			observability.String("messaging.system", "rabbitmq"),
			observability.String("messaging.destination", routingKey),
			observability.String("messaging.message_id", messageID),
		)
	})
	if errors.Is(err, circuitbreaker.ErrOpen) {
		err = fmt.Errorf("%w: %s", model.ErrUnavailable, breakerName)
	}
	if err != nil {
		status = "error"
	}

	publisher.kit.Metrics().RecordRabbitMQPublished(ctx, routingKey, status)
	return err
}

func (publisher *Publisher) Ping(ctx context.Context) error {
	return publisher.client.Ping(ctx)
}

func (publisher *Publisher) Close() error {
	return publisher.client.Close()
}
