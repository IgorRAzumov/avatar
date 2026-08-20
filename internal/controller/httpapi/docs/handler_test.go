package docs_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/docs"

	"github.com/stretchr/testify/assert"
)

func TestSpecServesOpenAPI(t *testing.T) {
	rec := httptest.NewRecorder()
	docs.New().Spec(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/yaml", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "openapi: 3.0.3")
	assert.Contains(t, rec.Body.String(), "/api/v1/avatars")
}

func TestUIServesSwagger(t *testing.T) {
	rec := httptest.NewRecorder()
	docs.New().UI(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "swagger-ui")
	assert.Contains(t, rec.Body.String(), `url: "/openapi.yaml"`)
}
