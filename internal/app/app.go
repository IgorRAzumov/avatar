package app

import (
	"context"
	"net/http"

	"avatar/internal/adapter/postgres"
	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/config"
	"avatar/internal/controller/httpapi"
	apimiddleware "avatar/internal/controller/httpapi/common/middleware"
	"avatar/internal/controller/httpapi/docs"
	"avatar/internal/controller/httpapi/health"
	"avatar/internal/controller/httpapi/read"
	"avatar/internal/controller/httpapi/web"
	"avatar/internal/controller/httpapi/write"
	domainrepo "avatar/internal/domain/repository"
	domainhealth "avatar/internal/domain/usecase/health"
	domainread "avatar/internal/domain/usecase/read"
	domainwrite "avatar/internal/domain/usecase/write"
	"avatar/internal/logger"
	"avatar/internal/observability"
)

func Run(log *logger.Logger, appConfig *config.Config, kit observability.Kit, metricsHandler http.Handler) error {
	serviceLog := log.WithService(appConfig.Observability.ServiceName)

	avatarStore, closeStore, err := postgres.Open(context.Background(), appConfig.Postgres, kit)
	if err != nil {
		return err
	}
	defer closeStore()

	fileStore, err := NewFileStore(appConfig, kit)
	if err != nil {
		return err
	}

	publisher, broker, err := NewPublisher(appConfig, avatarStore, avatarStore, fileStore, serviceLog, kit)
	if err != nil {
		return err
	}
	if closer, ok := publisher.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				serviceLog.Error(context.Background(), "close publisher", "error", err)
			}
		}()
	}

	router := newRouter(appConfig, avatarStore, fileStore, publisher, broker, serviceLog, kit)
	return httpapi.StartServer(serviceLog, appConfig, router, metricsHandler)
}

func newRouter(
	cfg *config.Config,
	avatarStore *repository.PostgresAvatarStore,
	filesStorage domainrepo.FilesRepository,
	publisher domainrepo.EventPublisherRepository,
	broker Pinger,
	log *logger.Logger,
	kit observability.Kit,
) http.Handler {
	rawRead := domainread.NewReadUsecase(avatarStore, filesStorage)
	avatarQuery := newInstrumentedReadUsecase(rawRead, kit)
	avatarCommand := newInstrumentedWriteUsecase(
		domainwrite.NewWriteUsecase(
			avatarStore,
			avatarStore,
			filesStorage,
			publisher,
			cfg.MaxUploadBytes(),
		),
		rawRead,
		kit,
	)
	healthChecker := domainhealth.NewChecker(avatarStore, filesStorage, "s3", broker)

	avatarWrite := write.New(log, avatarCommand, cfg.MaxUploadBytes(), cfg.Server.BaseURL)

	return httpapi.NewRouter(httpapi.RouterDeps{
		Logger:      log,
		ServiceName: cfg.Observability.ServiceName,
		Kit:         kit,
		AvatarRead:  read.New(log, avatarQuery, cfg.Server.BaseURL),
		AvatarWrite: avatarWrite,
		Web:         web.New(avatarWrite),
		Docs:        docs.New(),
		Health:      health.New(healthChecker),
		RateLimit: apimiddleware.RateLimitSettings{
			Enabled: cfg.RateLimit.Enabled,
			RPS:     cfg.RateLimit.RPS,
		},
		TrustedProxyCIDRs: cfg.Server.TrustedProxyCIDRs,
	})
}
