package app

import (
	"avatar/internal/observability"
	"context"
	"net/http"
	"os"

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

func InitApp(cfg *config.Config, serviceName string) (*logger.Logger, *observability.Runtime, error) {
	obs, err := initObservability(cfg, serviceName)
	if err != nil {
		return nil, nil, err
	}
	log := initLogger(cfg, obs, serviceName)
	return log, &obs, nil
}

func Run(log *logger.Logger, appConfig *config.Config, kit observability.Kit) error {
	serviceLog := log.WithService(appConfig.Observability.ServiceName)

	avatarStore, closeStore, err := postgres.Open(context.Background(), appConfig.Postgres.DSN, kit)
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

	return httpapi.StartServer(serviceLog, appConfig, newRouter(appConfig, avatarStore, fileStore, publisher, broker, serviceLog, kit))
}

func initLogger(cfg *config.Config, obs observability.Runtime, serviceName string) *logger.Logger {
	return logger.New(
		os.Stdout,
		logger.ParseLevel(cfg.Observability.LogLevel),
		logger.WithOTLP(obs.LogProvider, serviceName),
	).WithService(serviceName)
}

func initObservability(cfg *config.Config, serviceName string) (observability.Runtime, error) {
	obs, err := observability.Init(context.Background(), observability.Config{
		ServiceName:      serviceName,
		ServiceVersion:   cfg.Observability.ServiceVersion,
		Environment:      cfg.Observability.Environment,
		TracingEnabled:   cfg.Observability.TracingEnabled,
		LogsEnabled:      cfg.Observability.LogsEnabled,
		MetricsEnabled:   cfg.Observability.MetricsEnabled,
		OTLPEndpoint:     cfg.Observability.OTLPEndpoint,
		TraceSampleRatio: cfg.Observability.TraceSampleRatio,
	})
	if err != nil {
		return observability.Runtime{}, err
	}
	return obs, nil
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
		Health:      health.New(healthChecker),
	})
}
