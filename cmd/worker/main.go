package main

import (
	"avatar/cmd"
	"context"
	"errors"
	"os/signal"
	"syscall"

	"avatar/internal/app"
	"avatar/internal/config"
	"avatar/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}

	serviceName := cfg.Observability.ServiceName + "-worker"

	log, runtime, err := app.InitApp(cfg, serviceName)
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}
	defer func() { cmd.ShutdownRuntime(runtime) }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := worker.Run(ctx, log, cfg); err != nil && !errors.Is(err, context.Canceled) {
		cmd.ShutdownRuntime(runtime)
		cmd.ExitWithLog(log, err)
	}
}
