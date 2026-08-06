package worker

import (
	"context"
	"fmt"

	"avatar/internal/adapter/postgres"
	"avatar/internal/adapter/rabbitmq"
	"avatar/internal/app"
	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/processor"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config, kit observability.Kit) error {
	if !cfg.RabbitMQ.Enabled {
		return fmt.Errorf("worker requires RABBITMQ_ENABLED=true")
	}

	avatarStore, closeStore, err := postgres.Open(ctx, cfg.Postgres.DSN, kit)
	if err != nil {
		return err
	}
	defer closeStore()

	fileStore, err := app.NewFileStore(cfg, kit)
	if err != nil {
		return err
	}

	imageProcessor := processor.NewImageProcessor(avatarStore, avatarStore, fileStore, kit)
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ, imageProcessor, log, kit)
	if err != nil {
		return err
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error(ctx, "close consumer", "error", err)
		}
	}()

	log.Info(ctx, "worker consuming events",
		"exchange", cfg.RabbitMQ.Exchange,
		"upload_queue", config.DefaultRabbitMQUploadQueue,
		"delete_queue", config.DefaultRabbitMQDeleteQueue,
	)

	return consumer.Run(ctx)
}
