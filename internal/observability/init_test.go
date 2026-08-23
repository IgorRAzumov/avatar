package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/observability"

	"github.com/stretchr/testify/require"
)

func TestInitPrometheusHandlerExposesMetrics(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		MetricsEnabled: true,
	})
	require.NoError(t, err)
	require.NotNil(t, runtime.MetricsHandler)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })

	runtime.Kit.Metrics().RecordHTTPRequest(context.Background(), "GET", "/api/v1/avatars", "200", 0)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	runtime.MetricsHandler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "http_requests_total")
}

func TestInitAllObservabilityEnabled(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:      "test",
		ServiceVersion:   "1.0.0",
		Environment:      "test",
		TracingEnabled:   true,
		LogsEnabled:      true,
		MetricsEnabled:   true,
		OTLPEndpoint:     "localhost:4317",
		TraceSampleRatio: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, runtime.Kit)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })
}
