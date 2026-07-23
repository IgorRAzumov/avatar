package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedStaticFiles(t *testing.T) {
	fsys, err := web.StaticFS()
	require.NoError(t, err)

	for _, name := range []string{"upload.html", "gallery.html"} {
		_, err := fsys.Open(name)
		assert.NoError(t, err, name)
	}
}

func TestUploadPage(t *testing.T) {
	h := web.New(nil)
	req := httptest.NewRequest(http.MethodGet, "/web/upload", nil)
	rec := httptest.NewRecorder()

	h.UploadPage(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "GophProfile")
}

func TestGalleryPage(t *testing.T) {
	h := web.New(nil)
	req := httptest.NewRequest(http.MethodGet, "/web/gallery/user@test.com", nil)
	rec := httptest.NewRecorder()

	h.GalleryPage(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Галерея")
}
