package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"avatar/cmd"
	"avatar/internal/config"
	"avatar/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}

	log, runtime, err := cmd.InitApp(cfg, cfg.Observability.ServiceName)
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}
	defer func() { cmd.ShutdownRuntime(runtime) }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = worker.Run(ctx, log, cfg, runtime.Kit, runtime.MetricsHandler)
	if err != nil && !errors.Is(err, context.Canceled) {
		cmd.ShutdownRuntime(runtime)
		cmd.ExitWithLog(log, err)
	}
}
