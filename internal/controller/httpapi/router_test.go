package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/health"
	"avatar/internal/controller/httpapi/web"
	"avatar/internal/controller/httpapi/write"
	domainhealth "avatar/internal/domain/usecase/health"
	"avatar/internal/logger"
	"avatar/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterHealthRoute(t *testing.T) {
	checker := domainhealth.NewChecker(nil, nil, "storage", nil)
	router := NewRouter(RouterDeps{
		Logger:      logger.Nop(),
		ServiceName: "avatar-service-test",
		Kit:         observability.NewTestKit(),
		Health:      health.New(checker),
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestRouterAPIRoutesRegistered(t *testing.T) {
	checker := domainhealth.NewChecker(nil, nil, "storage", nil)

	router := NewRouter(RouterDeps{
		Logger:      logger.Nop(),
		ServiceName: "avatar-service-test",
		Kit:         observability.NewTestKit(),
		AvatarWrite: write.New(logger.Nop(), nil, 1024, "http://localhost"),
		Health:      health.New(checker),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRouterWebUploadPage(t *testing.T) {
	writeHandler := write.New(logger.Nop(), nil, 1024, "http://localhost")
	router := NewRouter(RouterDeps{
		Logger:      logger.Nop(),
		ServiceName: "avatar-service-test",
		Kit:         observability.NewTestKit(),
		AvatarWrite: writeHandler,
		Web:         web.New(writeHandler),
	})

	req := httptest.NewRequest(http.MethodGet, "/web/upload", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/web/upload", rec.Header().Get("Location"))
}
