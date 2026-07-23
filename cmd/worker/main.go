package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/worker"
)

func main() {
	log := logger.New(os.Stdout)
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := worker.Run(ctx, log, cfg); err != nil && err != context.Canceled {
		log.Error("worker error", "error", err)
		os.Exit(1)
	}

	log.Info("worker stopped")
}
