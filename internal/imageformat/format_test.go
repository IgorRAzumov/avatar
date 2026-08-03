package imageformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedMimeTypes(t *testing.T) {
	assert.True(t, IsSupportedMimeType(MIMEJPEG))
	assert.True(t, IsSupportedMimeType(MIMEPNG))
	assert.True(t, IsSupportedMimeType(MIMEWebP))
	assert.False(t, IsSupportedMimeType("text/plain"))
	assert.Equal(t, JPEG, MimeToFormat(MIMEJPEG))
	assert.Equal(t, PNG, MimeToFormat(MIMEPNG))
	assert.Equal(t, WebP, MimeToFormat(MIMEWebP))
}

func TestFormatToMime(t *testing.T) {
	assert.Equal(t, MIMEJPEG, FormatToMime("jpg"))
	assert.Equal(t, MIMEPNG, FormatToMime(PNG))
	assert.Equal(t, MIMEWebP, FormatToMime(WebP))
}

func TestMIMEFromFileName(t *testing.T) {
	assert.Equal(t, MIMEJPEG, MIMEFromFileName("photo.jpg"))
	assert.Equal(t, MIMEPNG, MIMEFromFileName("photo.png"))
	assert.Equal(t, "", MIMEFromFileName("photo.txt"))
}

func TestSupportedFormatsLabel(t *testing.T) {
	assert.Equal(t, "jpeg, png, webp", SupportedFormatsLabel())
}

func TestDetectMimeType(t *testing.T) {
	assert.Equal(t, MIMEPNG, DetectMimeType([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}))
	assert.Equal(t, MIMEJPEG, DetectMimeType([]byte{0xFF, 0xD8, 0xFF, 0xE0}))
	assert.Equal(t, "", DetectMimeType([]byte("this is plain text, not an image")))
	assert.Equal(t, "", DetectMimeType(nil))
}
