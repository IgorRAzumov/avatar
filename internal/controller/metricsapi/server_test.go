package metricsapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerServesScrapeEndpoint(t *testing.T) {
	metrics := New(logger.Nop(), ":9090", http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte("# TYPE up gauge\nup 1\n"))
		}))
	require.NotNil(t, metrics)

	rec := httptest.NewRecorder()
	metrics.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "up 1")

	rec = httptest.NewRecorder()
	metrics.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestServerDisabledIsNoop(t *testing.T) {
	assert.Nil(t, New(logger.Nop(), ":9090", nil))
	assert.Nil(t, New(logger.Nop(), "", http.NewServeMux()))

	var metrics *Server
	metrics.Start()
	metrics.Shutdown(context.Background())
}
