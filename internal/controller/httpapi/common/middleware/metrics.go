package middleware

import (
	"net/http"
	"strconv"
	"time"

	"avatar/internal/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func PrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if ShouldSkipObservability(request.URL.Path) {
			next.ServeHTTP(writer, request)
			return
		}

		start := time.Now()
		wrapResponseWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
		next.ServeHTTP(wrapResponseWriter, request)

		status := strconv.Itoa(wrapResponseWriter.Status())
		path := routePattern(request)

		observability.RecordHTTPRequest(request.Method, path, status, time.Since(start))
	})
}

func routePattern(request *http.Request) string {
	if routeContext := chi.RouteContext(request.Context()); routeContext != nil {
		if pattern := routeContext.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return request.URL.Path
}
