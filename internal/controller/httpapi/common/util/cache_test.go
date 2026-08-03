package util_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/imageformat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageETagStable(t *testing.T) {
	data := []byte("image-bytes")
	a := util.ImageETag(data)
	b := util.ImageETag(data)
	assert.Equal(t, a, b)
	assert.True(t, len(a) > 2)
	assert.Equal(t, '"', rune(a[0]))
	assert.Equal(t, '"', rune(a[len(a)-1]))
}

func TestWriteCachedImageNotModified(t *testing.T) {
	data := []byte("jpeg-data")
	etag := util.ImageETag(data)

	req := httptest.NewRequest(http.MethodGet, "/img", nil)
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()

	util.WriteCachedImage(rec, req, data, imageformat.MIMEJPEG, "max-age=86400")

	assert.Equal(t, http.StatusNotModified, rec.Code)
	assert.Equal(t, etag, rec.Header().Get("ETag"))
	assert.Empty(t, rec.Body.Bytes())
}

func TestWriteCachedImageOK(t *testing.T) {
	data := []byte("jpeg-data")

	req := httptest.NewRequest(http.MethodGet, "/img", nil)
	rec := httptest.NewRecorder()

	util.WriteCachedImage(rec, req, data, imageformat.MIMEJPEG, "max-age=86400")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, data, rec.Body.Bytes())
	assert.NotEmpty(t, rec.Header().Get("ETag"))
	assert.Equal(t, "max-age=86400", rec.Header().Get("Cache-Control"))
}
