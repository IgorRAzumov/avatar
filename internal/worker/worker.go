package worker

import (
	"context"
	"fmt"
	"net/http"

	"avatar/internal/adapter/postgres"
	"avatar/internal/adapter/rabbitmq"
	"avatar/internal/adapter/s3"
	"avatar/internal/config"
	"avatar/internal/controller/metricsapi"
	"avatar/internal/domain/usecase/process"
	"avatar/internal/logger"
	"avatar/internal/observability"
)

func Run(
	ctx context.Context,
	log *logger.Logger,
	cfg *config.Config,
	kit observability.Kit,
	metricsHandler http.Handler,
) error {
	if !cfg.RabbitMQ.Enabled {
		return fmt.Errorf("worker requires RABBITMQ_ENABLED=true")
	}

	metrics := metricsapi.New(log, cfg.Server.MetricsAddr, metricsHandler)
	metrics.Start()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.DefaultShutdownTimeout)
		defer cancel()
		metrics.Shutdown(shutdownCtx)
	}()

	avatarStore, closeStore, err := postgres.Open(ctx, cfg.Postgres, kit)
	if err != nil {
		return err
	}
	defer closeStore()

	fileStore, err := s3.NewStorage(cfg.S3, kit)
	if err != nil {
		return err
	}

	imageProcessor := process.NewImageProcessor(avatarStore, avatarStore, fileStore, kit)
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
