package rabbitmq

import (
	"context"
	"fmt"

	"avatar/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
}

func NewClient(cfg config.RabbitMQConfig) (*Client, error) {
	connection, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	client := &Client{
		connection: connection,
		channel:    channel,
		exchange:   cfg.Exchange,
	}

	if err := client.declareTopology(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func (client *Client) declareTopology() error {
	if err := client.channel.ExchangeDeclare(
		client.exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	queues := []struct {
		name       string
		routingKey string
	}{
		{config.DefaultRabbitMQUploadQueue, config.DefaultRabbitMQUploadRoutingKey},
		{config.DefaultRabbitMQDeleteQueue, config.DefaultRabbitMQDeleteRoutingKey},
	}

	for _, queue := range queues {
		if _, err := client.channel.QueueDeclare(
			queue.name,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("declare queue %q: %w", queue.name, err)
		}
		if err := client.channel.QueueBind(
			queue.name,
			queue.routingKey,
			client.exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("bind queue %q: %w", queue.name, err)
		}
	}

	return nil
}

func (client *Client) Ping(_ context.Context) error {
	if client.connection == nil || client.connection.IsClosed() {
		return fmt.Errorf("rabbitmq connection closed")
	}
	return nil
}

func (client *Client) Close() error {
	if client.channel != nil {
		_ = client.channel.Close()
	}
	if client.connection != nil {
		return client.connection.Close()
	}
	return nil
}
