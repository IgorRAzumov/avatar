package middlewear_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/common/middlewear"
	"avatar/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf)

	called := false
	handler := middlewear.RequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
