package app

import (
	"context"
	"fmt"

	"avatar/internal/adapter/rabbitmq"
	"avatar/internal/config"
	domainrepo "avatar/internal/domain/repository"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/processor"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewPublisher(
	cfg *config.Config,
	readRepository domainrepo.ReadRepository,
	writeRepository domainrepo.WriteRepository,
	filesRepository domainrepo.FilesRepository,
	log *logger.Logger,
	kit observability.Kit,
) (domainrepo.EventPublisherRepository, Pinger, error) {
	imageProcessor := processor.NewImageProcessor(readRepository, writeRepository, filesRepository, kit)

	if cfg.RabbitMQ.Enabled {
		publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQ, kit)
		if err != nil {
			return nil, nil, fmt.Errorf("create rabbitmq publisher: %w", err)
		}
		return publisher, publisher, nil
	}

	asyncPublisher := processor.NewAsyncPublisher(imageProcessor, log, kit)
	return asyncPublisher, nil, nil
}
