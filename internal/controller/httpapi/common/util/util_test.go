package util_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	commonmodel "avatar/internal/controller/httpapi/common/model"
	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/domain/model"
	"avatar/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	util.WriteJSON(rec, http.StatusCreated, map[string]string{"a": "b"})

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "b", body["a"])
}

func TestWriteServiceErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", model.ErrNotFound, http.StatusNotFound},
		{"forbidden", model.ErrForbidden, http.StatusForbidden},
		{"invalid format", model.ErrInvalidFormat, http.StatusBadRequest},
		{"too large", model.ErrFileTooLarge, http.StatusRequestEntityTooLarge},
		{"unavailable", model.ErrUnavailable, http.StatusServiceUnavailable},
		{"internal", assert.AnError, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			util.WriteServiceError(context.Background(), logger.Nop(), rec, tc.err, 1024)
			assert.Equal(t, tc.want, rec.Code)

			var body commonmodel.ErrorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.NotEmpty(t, body.Error)
		})
	}
}

func TestAvatarPaths(t *testing.T) {
	assert.Equal(t, "/api/v1/avatars/abc", util.AvatarRelativePath("abc"))
	assert.Equal(t, "http://host/api/v1/avatars/abc", util.AvatarURL("http://host/", "abc"))
	assert.Equal(t,
		"http://host/api/v1/avatars/abc?size=100x100&format=webp",
		util.AvatarURLWithParams("http://host", "abc", "100x100", "webp"),
	)
	assert.Equal(t, "http://host/api/v1/avatars/abc", util.AvatarURLWithParams("http://host", "abc", "", ""))
}

func TestPathUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1itcode1%40gmail.com/avatars", nil)
	req.SetPathValue("user_id", "1itcode1%40gmail.com")
	assert.Equal(t, "1itcode1@gmail.com", util.PathUserID(req))

	req.SetPathValue("user_id", "user@example.com")
	assert.Equal(t, "user@example.com", util.PathUserID(req))
}
