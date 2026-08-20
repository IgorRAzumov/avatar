package main

import (
	"context"
	"os"

	"avatar/cmd"
	"avatar/internal/adapter/postgres"
	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/retry"
)

const serviceName = "avatar-migrate"

func main() {
	cfg, err := config.Load()
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}

	log := logger.New(os.Stdout, logger.ParseLevel(cfg.Observability.LogLevel)).WithService(serviceName)
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultMigrateTimeout)
	defer cancel()

	err = retry.WithBackoff(ctx, retry.DefaultMaxAttempts,
		func() error { return postgres.Migrate(ctx, cfg.Postgres.DSN) },
		func(attempt int, err error) {
			log.Warn(ctx, "migrate attempt failed", "attempt", attempt, "error", err)
		},
	)
	if err != nil {
		cmd.ExitWithLog(log, err)
	}

	log.Info(ctx, "migrations applied")
}
