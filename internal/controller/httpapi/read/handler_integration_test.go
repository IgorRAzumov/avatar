package read_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/read/model"

	"avatar/internal/controller/httpapi/read"
	domainread "avatar/internal/domain/usecase/read"
	domainwrite "avatar/internal/domain/usecase/write"
	"avatar/internal/imageformat"
	"avatar/internal/logger"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReadHandler() (*read.Handler, *domainread.Usecase, *domainwrite.Usecase) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	return read.New(logger.Nop(), query, "http://localhost:8080"), query, command
}

func TestGetAvatarSuccess(t *testing.T) {
	h, _, command := newReadHandler()
	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(80, 60))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+resp.ID, nil)
	req.SetPathValue("avatar_id", resp.ID)
	rec := httptest.NewRecorder()
	h.GetAvatar(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, imageformat.MIMEJPEG, rec.Header().Get("Content-Type"))
	assert.NotEmpty(t, rec.Header().Get("ETag"))
	assert.NotEmpty(t, rec.Body.Bytes())

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+resp.ID, nil)
	req2.SetPathValue("avatar_id", resp.ID)
	req2.Header.Set("If-None-Match", rec.Header().Get("ETag"))
	rec2 := httptest.NewRecorder()
	h.GetAvatar(rec2, req2)
	assert.Equal(t, http.StatusNotModified, rec2.Code)
	assert.Empty(t, rec2.Body.Bytes())
}

func TestGetAvatarNotFound(t *testing.T) {
	h, _, _ := newReadHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/missing", nil)
	req.SetPathValue("avatar_id", "missing")
	rec := httptest.NewRecorder()
	h.GetAvatar(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetMetadataSuccess(t *testing.T) {
	h, _, command := newReadHandler()
	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(50, 50))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/"+resp.ID+"/metadata", nil)
	req.SetPathValue("avatar_id", resp.ID)
	rec := httptest.NewRecorder()
	h.GetMetadata(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var meta model.AvatarMetadataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &meta))
	assert.Equal(t, resp.ID, meta.ID)
}

func TestListUserAvatars(t *testing.T) {
	h, _, command := newReadHandler()
	_, err := command.Upload(context.Background(), "user@test.com", "one.jpg", testutil.JPEG(20, 20))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user@test.com/avatars", nil)
	req.SetPathValue("user_id", "user@test.com")
	rec := httptest.NewRecorder()
	h.GetListUserAvatars(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string][]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Len(t, body["avatars"], 1)
}

func TestListUserAvatarsEncodedEmail(t *testing.T) {
	h, _, command := newReadHandler()
	email := "1itcode1@gmail.com"
	_, err := command.Upload(context.Background(), email, "one.jpg", testutil.JPEG(20, 20))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1itcode1%40gmail.com/avatars", nil)
	req.SetPathValue("user_id", "1itcode1%40gmail.com")
	rec := httptest.NewRecorder()
	h.GetListUserAvatars(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string][]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Len(t, body["avatars"], 1)
}

func TestGetUserAvatar(t *testing.T) {
	h, _, command := newReadHandler()
	_, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(30, 30))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user@test.com/avatar", nil)
	req.SetPathValue("user_id", "user@test.com")
	rec := httptest.NewRecorder()
	h.GetUserAvatar(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListUserAvatarsEmpty(t *testing.T) {
	h, _, _ := newReadHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/empty@test.com/avatars", nil)
	req.SetPathValue("user_id", "empty@test.com")
	rec := httptest.NewRecorder()
	h.GetListUserAvatars(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUserAvatarPlaceholder(t *testing.T) {
	h, _, _ := newReadHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/nobody@test.com/avatar", nil)
	req.SetPathValue("user_id", "nobody@test.com")
	rec := httptest.NewRecorder()
	h.GetUserAvatar(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, imageformat.MIMEWebP, rec.Header().Get("Content-Type"))
	assert.NotEmpty(t, rec.Body.Bytes())
}
