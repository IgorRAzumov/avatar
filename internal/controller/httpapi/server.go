package httpapi

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"avatar/internal/config"
	"avatar/internal/logger"
)

func StartServer(log *logger.Logger, cfg *config.Config, handler http.Handler) error {
	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var serveErr error
	select {
	case <-stop:
		log.Info("server shutting down")
	case err := <-errCh:
		log.Error("server error", "error", err)
		serveErr = err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.DefaultShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
		return err
	}
	log.Info("server stopped")
	return serveErr
}
