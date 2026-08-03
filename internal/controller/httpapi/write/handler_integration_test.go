package write_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/config"
	"avatar/internal/controller/httpapi/write"
	domainwrite "avatar/internal/domain/usecase/write"
	"avatar/internal/logger"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWriteHandler() (*write.Handler, *domainwrite.Usecase) {
	_, command, _, _ := testutil.NewAvatarUseCases()
	return write.New(logger.Nop(), command, config.DefaultMaxUploadBytes(), "http://localhost:8080"), command
}

func makeJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestUploadSuccess(t *testing.T) {
	h, _ := newWriteHandler()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "photo.jpg")
	require.NoError(t, err)
	_, err = part.Write(makeJPEG(40, 40))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(write.HeaderUserID, "user@test.com")
	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestDeleteAvatarForbidden(t *testing.T) {
	h, command := newWriteHandler()
	resp, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", makeJPEG(20, 20))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+resp.ID, nil)
	req.SetPathValue("avatar_id", resp.ID)
	req.Header.Set(write.HeaderUserID, "other@test.com")
	rec := httptest.NewRecorder()
	h.DeleteAvatar(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDeleteAvatarSuccess(t *testing.T) {
	h, command := newWriteHandler()
	resp, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", makeJPEG(20, 20))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/"+resp.ID, nil)
	req.SetPathValue("avatar_id", resp.ID)
	req.Header.Set(write.HeaderUserID, "owner@test.com")
	rec := httptest.NewRecorder()
	h.DeleteAvatar(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteUserAvatarSuccess(t *testing.T) {
	h, command := newWriteHandler()
	_, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", makeJPEG(20, 20))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/owner@test.com/avatar", nil)
	req.SetPathValue("user_id", "owner@test.com")
	req.Header.Set(write.HeaderUserID, "owner@test.com")
	rec := httptest.NewRecorder()
	h.DeleteUserAvatar(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUploadInvalidFormat(t *testing.T) {
	h, _ := newWriteHandler()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "notes.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(write.HeaderUserID, "user@test.com")
	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
