package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"avatar/internal/imageformat"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type Resizer struct{}

func NewResizer() *Resizer {
	return &Resizer{}
}

func (resizer *Resizer) Decode(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image: %w", err)
	}
	return img, format, nil
}

func (resizer *Resizer) Dimensions(data []byte) (width, height int, err error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("decode config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

func (resizer *Resizer) Resize(data []byte, width, height int) ([]byte, error) {
	img, format, err := resizer.Decode(data)
	if err != nil {
		return nil, err
	}

	resized := imaging.Fit(img, width, height, imaging.Lanczos)
	return encodeImage(resized, format)
}

// ResizeAsJPEG always encodes thumbnails as JPEG regardless of the source format.
func (resizer *Resizer) ResizeAsJPEG(data []byte, width, height int) ([]byte, error) {
	img, _, err := resizer.Decode(data)
	if err != nil {
		return nil, err
	}

	resized := imaging.Fit(img, width, height, imaging.Lanczos)
	return encodeImage(resized, imageformat.JPEG)
}

func (resizer *Resizer) Convert(data []byte, targetFormat string) ([]byte, string, error) {
	img, _, err := resizer.Decode(data)
	if err != nil {
		return nil, "", err
	}

	encoded, err := encodeImage(img, targetFormat)
	if err != nil {
		return nil, "", err
	}

	return encoded, imageformat.FormatToMime(targetFormat), nil
}

func encodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch imageformat.NormalizeFormat(format) {
	case imageformat.PNG:
		err = imaging.Encode(&buf, img, imaging.PNG)
	default:
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(imageformat.JPEGQuality))
	}

	if err != nil {
		return nil, fmt.Errorf("encode image: %w", err)
	}
	return buf.Bytes(), nil
}

func DetectMimeType(data []byte) string {
	return imageformat.DetectMimeType(data)
}
