package model

const (
	ImageSizeOriginal = "original"
	ImageSizeSmall    = "100x100"
	ImageSizeMedium   = "300x300"
)

const DefaultPlaceholderSide = 300

type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

var ThumbnailSizes = []ThumbnailSize{
	{Name: ImageSizeSmall, Width: 100, Height: 100},
	{Name: ImageSizeMedium, Width: 300, Height: 300},
}

func IsOriginalImageSize(size string) bool {
	return size == "" || size == ImageSizeOriginal
}

func LookupThumbnailSize(size string) (ThumbnailSize, bool) {
	for _, thumbnail := range ThumbnailSizes {
		if thumbnail.Name == size {
			return thumbnail, true
		}
	}
	return ThumbnailSize{}, false
}

func PlaceholderDimensions(size string) (width, height int) {
	if thumbnail, ok := LookupThumbnailSize(size); ok {
		return thumbnail.Width, thumbnail.Height
	}
	return DefaultPlaceholderSide, DefaultPlaceholderSide
}
