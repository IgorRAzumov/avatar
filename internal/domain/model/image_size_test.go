package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupThumbnailSize(t *testing.T) {
	thumbnail, ok := LookupThumbnailSize(ImageSizeSmall)
	assert.True(t, ok)
	assert.Equal(t, 100, thumbnail.Width)

	_, ok = LookupThumbnailSize("999x999")
	assert.False(t, ok)
}

func TestIsOriginalImageSize(t *testing.T) {
	assert.True(t, IsOriginalImageSize(""))
	assert.True(t, IsOriginalImageSize(ImageSizeOriginal))
	assert.False(t, IsOriginalImageSize(ImageSizeSmall))
}

func TestPlaceholderDimensions(t *testing.T) {
	w, h := PlaceholderDimensions(ImageSizeSmall)
	assert.Equal(t, 100, w)
	assert.Equal(t, 100, h)

	w, h = PlaceholderDimensions(ImageSizeMedium)
	assert.Equal(t, 300, w)
	assert.Equal(t, 300, h)

	w, h = PlaceholderDimensions(ImageSizeOriginal)
	assert.Equal(t, DefaultPlaceholderSide, w)
	assert.Equal(t, DefaultPlaceholderSide, h)
}
