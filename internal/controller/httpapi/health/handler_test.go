package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/health"
	"avatar/internal/controller/httpapi/health/model"
	domainhealth "avatar/internal/domain/usecase/health"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	checker := domainhealth.NewChecker(nil, nil, "storage", nil)
	h := health.New(checker)

	req := httptest.NewRequest(http.MethodGet, "/healthRepository", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp model.HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
}
