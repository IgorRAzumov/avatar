package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"avatar/internal/config"
	"avatar/internal/domain/model"

	amqp "github.com/rabbitmq/amqp091-go"
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
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

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
		},
	)
}

func (publisher *Publisher) Ping(ctx context.Context) error {
	return publisher.client.Ping(ctx)
}

func (publisher *Publisher) Close() error {
	return publisher.client.Close()
}
