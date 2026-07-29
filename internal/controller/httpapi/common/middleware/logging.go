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
			start := time.Now()
			wrapResponseWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			next.ServeHTTP(wrapResponseWriter, request)

			args := []any{
				"method", request.Method,
				"path", request.URL.Path,
				"status", wrapResponseWriter.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			}
			if requestID := middleware.GetReqID(request.Context()); requestID != "" {
				args = append(args, "request_id", requestID)
			}
			logger.Info("request", args...)
		})
	}
}
