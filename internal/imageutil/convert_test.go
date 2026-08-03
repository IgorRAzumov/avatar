package imageutil_test

import (
	"testing"

	"avatar/internal/imageformat"
	"avatar/internal/imageutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToPNG(t *testing.T) {
	data := makeJPEG(20, 20)
	r := imageutil.NewResizer()

	converted, mimeType, err := r.Convert(data, imageformat.PNG)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEPNG, mimeType)
	assert.NotEmpty(t, converted)
}

func TestConvertInvalidImage(t *testing.T) {
	r := imageutil.NewResizer()
	_, _, err := r.Convert([]byte("bad"), imageformat.JPEG)
	require.Error(t, err)
}

func TestDetectMimeTypes(t *testing.T) {
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	assert.Equal(t, imageformat.MIMEPNG, imageutil.DetectMimeType(pngHeader))
}
