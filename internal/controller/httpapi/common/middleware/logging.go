package middleware

import (
	"net/http"
	"time"

	"avatar/internal/logger"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if ShouldSkip(request.URL.Path) {
				next.ServeHTTP(writer, request)
				return
			}

			start := time.Now()
			wrapResponseWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			next.ServeHTTP(wrapResponseWriter, request)

			status := wrapResponseWriter.Status()
			ctx := request.Context()
			args := []any{
				"method", request.Method,
				"path", routePattern(request),
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
			}

			if status >= 500 {
				logger.Error(ctx, "request", args...)
			} else if status >= 400 {
				logger.Warn(ctx, "request", args...)
			} else {
				logger.Info(ctx, "request", args...)
			}
		})
	}
}
