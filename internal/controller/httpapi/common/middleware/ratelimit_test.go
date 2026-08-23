package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/common/middleware"
	commonmodel "avatar/internal/controller/httpapi/common/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitDisabledIsNoop(t *testing.T) {
	calls := 0
	handler := middleware.RateLimit(middleware.RateLimitSettings{})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		calls++
		writer.WriteHeader(http.StatusNoContent)
	}))

	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
	assert.Equal(t, 5, calls)
}

func TestRateLimitAllowsConfiguredRPS(t *testing.T) {
	handler := middleware.RateLimit(middleware.RateLimitSettings{
		Enabled: true,
		RPS:     2,
	})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	for range 2 {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil))
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimitRejectsExcessRequests(t *testing.T) {
	handler := middleware.RateLimit(middleware.RateLimitSettings{
		Enabled: true,
		RPS:     1,
	})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil))
	assert.Equal(t, http.StatusNoContent, first.Code)

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil))
	assert.Equal(t, http.StatusTooManyRequests, second.Code)
	assert.Equal(t, "1", second.Header().Get("Retry-After"))

	var body commonmodel.ErrorResponse
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &body))
	assert.Equal(t, commonmodel.MsgTooManyRequests, body.Error)
}

func TestRateLimitIgnoresForgedForwardedForWithoutTrustedProxies(t *testing.T) {
	handler := middleware.ClientIP(nil)(middleware.RateLimit(middleware.RateLimitSettings{
		Enabled: true,
		RPS:     1,
	})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})))

	first := httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil)
	first.Header.Set("X-Forwarded-For", "1.1.1.1")
	firstRec := httptest.NewRecorder()
	handler.ServeHTTP(firstRec, first)
	assert.Equal(t, http.StatusNoContent, firstRec.Code)

	second := httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil)
	second.Header.Set("X-Forwarded-For", "2.2.2.2")
	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, second)
	assert.Equal(t, http.StatusTooManyRequests, secondRec.Code)
}

func TestRateLimitKeysByForwardedForBehindTrustedProxy(t *testing.T) {
	handler := middleware.ClientIP([]string{"192.0.2.0/24", ""})(middleware.RateLimit(middleware.RateLimitSettings{
		Enabled: true,
		RPS:     1,
	})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})))

	for _, clientAddress := range []string{"1.1.1.1", "2.2.2.2"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil)
		req.Header.Set("X-Forwarded-For", clientAddress)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code, clientAddress)
	}

	repeated := httptest.NewRequest(http.MethodGet, "/api/v1/avatars", nil)
	repeated.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, repeated)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimitSkipsInfrastructurePaths(t *testing.T) {
	handler := middleware.RateLimit(middleware.RateLimitSettings{
		Enabled: true,
		RPS:     1,
	})(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/health", "/health/live", "/openapi.yaml", "/docs"} {
		for range 3 {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			assert.Equal(t, http.StatusOK, rec.Code, path)
		}
	}
}
