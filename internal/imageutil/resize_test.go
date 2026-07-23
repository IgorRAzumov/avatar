package imageutil_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"avatar/internal/imageformat"
	"avatar/internal/imageutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestDetectMimeType(t *testing.T) {
	data := makeJPEG(10, 10)
	assert.Equal(t, imageformat.MIMEJPEG, imageutil.DetectMimeType(data))
	assert.Equal(t, "", imageutil.DetectMimeType([]byte("not-an-image")))
}

func TestResizerDimensionsAndResize(t *testing.T) {
	data := makeJPEG(800, 600)
	r := imageutil.NewResizer()

	w, h, err := r.Dimensions(data)
	require.NoError(t, err)
	assert.Equal(t, 800, w)
	assert.Equal(t, 600, h)

	resized, err := r.Resize(data, 100, 100)
	require.NoError(t, err)
	assert.NotEmpty(t, resized)

	rw, rh, err := r.Dimensions(resized)
	require.NoError(t, err)
	assert.LessOrEqual(t, rw, 100)
	assert.LessOrEqual(t, rh, 100)
}

func TestResizerResizeAsJPEGFromPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	resized, err := imageutil.NewResizer().ResizeAsJPEG(buf.Bytes(), 100, 100)
	require.NoError(t, err)
	assert.Equal(t, byte(0xFF), resized[0])
	assert.Equal(t, byte(0xD8), resized[1])
}
