package cmd

import (
	"avatar/internal/config"
	"context"
	"os"
	"time"

	"avatar/internal/logger"
	"avatar/internal/observability"
)

func InitApp(cfg *config.Config, serviceName string) (*logger.Logger, *observability.Runtime, error) {
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
		return nil, nil, err
	}
	log := logger.New(
		os.Stdout,
		logger.ParseLevel(cfg.Observability.LogLevel),
		logger.WithOTLP(obs.LogProvider, serviceName),
	).WithService(serviceName)
	return log, &obs, nil
}

func ShutdownRuntime(runtime *observability.Runtime) {
	if runtime == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runtime.Shutdown(ctx); err != nil {
		_, _ = os.Stderr.WriteString("shutdown observability: " + err.Error() + "\n")
	}
}

func ExitWithLog(log *logger.Logger, err error) {
	if err != nil {
		if log != nil {
			log.Error(context.Background(), "application error", "error", err)
		} else {
			_, _ = os.Stderr.WriteString("application error: " + err.Error() + "\n")
		}
	}
	os.Exit(1)
}
