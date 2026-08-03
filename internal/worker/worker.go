package worker

import (
	"context"
	"fmt"

	"avatar/internal/adapter/postgres"
	"avatar/internal/adapter/rabbitmq"
	"avatar/internal/app"
	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/processor"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	if !cfg.RabbitMQ.Enabled {
		return fmt.Errorf("worker requires RABBITMQ_ENABLED=true")
	}

	avatarStore, err := postgres.Open(ctx, cfg.Postgres.DSN)
	if err != nil {
		return err
	}
	defer avatarStore.Close()

	fileStore, err := app.NewFileStore(cfg)
	if err != nil {
		return err
	}

	imageProcessor := processor.NewImageProcessor(avatarStore, avatarStore, fileStore)
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ, imageProcessor, log)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("close consumer", "error", err)
		}
	}()

	log.Info("worker consuming events",
		"exchange", cfg.RabbitMQ.Exchange,
		"upload_queue", config.DefaultRabbitMQUploadQueue,
		"delete_queue", config.DefaultRabbitMQDeleteQueue,
	)

	return consumer.Run(ctx)
}
