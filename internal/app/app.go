package app

import (
	"context"
	"net/http"

	"avatar/internal/adapter/postgres"
	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/config"
	"avatar/internal/controller/httpapi"
	"avatar/internal/controller/httpapi/health"
	"avatar/internal/controller/httpapi/read"
	"avatar/internal/controller/httpapi/web"
	"avatar/internal/controller/httpapi/write"
	domainrepo "avatar/internal/domain/repository"
	domainhealth "avatar/internal/domain/usecase/health"
	domainread "avatar/internal/domain/usecase/read"
	domainwrite "avatar/internal/domain/usecase/write"
	"avatar/internal/logger"
)

func Run(logger *logger.Logger) error {
	appConfig, err := config.Load()
	if err != nil {
		return err
	}

	avatarStore, err := postgres.Open(context.Background(), appConfig.Postgres.DSN)
	if err != nil {
		return err
	}
	defer avatarStore.Close()

	fileStore, err := NewFileStore(appConfig)
	if err != nil {
		return err
	}

	publisher, broker, err := NewPublisher(appConfig, avatarStore, avatarStore, fileStore, logger)
	if err != nil {
		return err
	}
	if closer, ok := publisher.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				logger.Error("close publisher", "error", err)
			}
		}()
	}

	return httpapi.StartServer(logger, appConfig, newRouter(appConfig, avatarStore, fileStore, publisher, broker, logger))
}

func newRouter(
	cfg *config.Config,
	avatarStore *repository.PostgresAvatarStore,
	filesStorage domainrepo.FilesRepository,
	publisher domainrepo.EventPublisherRepository,
	broker Pinger,
	log *logger.Logger,
) http.Handler {
	avatarQuery := domainread.NewReadUsecase(avatarStore, filesStorage)
	avatarCommand := domainwrite.NewWriteUsecase(
		avatarStore,
		avatarStore,
		filesStorage,
		publisher,
		cfg.MaxUploadBytes(),
	)
	healthChecker := domainhealth.NewChecker(avatarStore, filesStorage, "s3", broker)

	avatarWrite := write.New(log, avatarCommand, cfg.MaxUploadBytes(), cfg.Server.BaseURL)

	return httpapi.NewRouter(httpapi.RouterDeps{
		Logger:      log,
		AvatarRead:  read.New(log, avatarQuery, cfg.Server.BaseURL),
		AvatarWrite: avatarWrite,
		Web:         web.New(avatarWrite),
		Health:      health.New(healthChecker),
	})
}
