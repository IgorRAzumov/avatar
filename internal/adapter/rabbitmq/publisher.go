package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/observability"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
)

type Publisher struct {
	client *Client
}

func NewPublisher(cfg config.RabbitMQConfig) (*Publisher, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Publisher{client: client}, nil
}

func (publisher *Publisher) PublishUploadEvent(ctx context.Context, event model.AvatarUploadEvent) error {
	return publisher.publish(ctx, config.DefaultRabbitMQUploadRoutingKey, event.AvatarID, event)
}

func (publisher *Publisher) PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error {
	return publisher.publish(ctx, config.DefaultRabbitMQDeleteRoutingKey, event.AvatarID, event)
}

func (publisher *Publisher) publish(ctx context.Context, routingKey, messageID string, payload any) error {
	status := "success"

	err := observability.Run(ctx, "rabbitmq.publish", func(ctx context.Context) error {
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
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", routingKey),
		attribute.String("messaging.message_id", messageID),
	)
	if err != nil {
		status = "error"
	}

	observability.RecordRabbitMQPublished(routingKey, status)
	return err
}

func (publisher *Publisher) Ping(ctx context.Context) error {
	return publisher.client.Ping(ctx)
}

func (publisher *Publisher) Close() error {
	return publisher.client.Close()
}
