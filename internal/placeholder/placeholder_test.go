package placeholder

import (
	"testing"

	"avatar/internal/domain/model"
	"avatar/internal/imageformat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageDefaultWebP(t *testing.T) {
	data, mimeType, err := Image("", "")
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEWebP, mimeType)
	assert.NotEmpty(t, data)
	assert.Equal(t, imageformat.MIMEWebP, imageformat.DetectMimeType(data))
}

func TestImageSizes(t *testing.T) {
	for _, size := range []string{model.ImageSizeSmall, model.ImageSizeMedium, model.ImageSizeOriginal} {
		data, mimeType, err := Image(size, "")
		require.NoError(t, err, size)
		assert.NotEmpty(t, data, size)
		assert.NotEmpty(t, mimeType, size)
	}
}

func TestImageJPEGFormat(t *testing.T) {
	data, mimeType, err := Image(model.ImageSizeSmall, imageformat.JPEG)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEJPEG, mimeType)
	assert.NotEmpty(t, data)
}

func TestImagePNGFormat(t *testing.T) {
	data, mimeType, err := Image(model.ImageSizeSmall, imageformat.PNG)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEPNG, mimeType)
	assert.NotEmpty(t, data)
}
