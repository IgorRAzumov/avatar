package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Tracing(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, serviceName,
			otelhttp.WithSpanNameFormatter(func(_ string, request *http.Request) string {
				return request.Method + " " + routePattern(request)
			}),
			otelhttp.WithFilter(func(request *http.Request) bool {
				return !ShouldSkip(request.URL.Path)
			}),
		)
	}
}
