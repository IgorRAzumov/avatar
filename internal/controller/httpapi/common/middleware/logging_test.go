package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/common/middleware"
	"avatar/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, slog.LevelInfo)

	called := false
	handler := middleware.RequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.True(t, called)
	assert.Equal(t, http.StatusTeapot, rec.Code)

	out := buf.String()
	assert.Contains(t, out, "request")
	assert.Contains(t, out, "GET")
	assert.Contains(t, out, "/some/path")
}
