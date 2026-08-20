package metricsapi

import (
	"context"
	"errors"
	"net/http"

	"avatar/internal/config"
	"avatar/internal/logger"
)

const scrapePath = "/metrics"

type Server struct {
	server *http.Server
	log    *logger.Logger
}

func New(log *logger.Logger, addr string, handler http.Handler) *Server {
	if handler == nil || addr == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(scrapePath, handler)

	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: config.DefaultServerReadTimeout,
		},
		log: log,
	}
}

func (metrics *Server) Start() {
	if metrics == nil {
		return
	}

	go func() {
		metrics.log.Info(context.Background(), "metrics server started", "addr", metrics.server.Addr)
		if err := metrics.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			metrics.log.Error(context.Background(), "metrics server error", "error", err)
		}
	}()
}

func (metrics *Server) Shutdown(ctx context.Context) {
	if metrics == nil {
		return
	}

	if err := metrics.server.Shutdown(ctx); err != nil {
		metrics.log.Error(ctx, "metrics server shutdown error", "error", err)
	}
}
