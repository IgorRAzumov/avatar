package placeholder

import (
	_ "embed"
	"fmt"

	"avatar/internal/domain/model"
	"avatar/internal/imageformat"
	"avatar/internal/imageutil"
)

//go:embed avatar.webp
var avatarWebP []byte

// Image returns the default avatar placeholder for users without an uploaded avatar.
func Image(size, format string) ([]byte, string, error) {
	resizer := imageutil.NewResizer()

	data := avatarWebP
	mimeType := imageformat.MIMEWebP

	if !model.IsOriginalImageSize(size) {
		width, height := model.PlaceholderDimensions(size)
		resized, err := resizer.Resize(data, width, height)
		if err != nil {
			return nil, "", fmt.Errorf("resize placeholder: %w", err)
		}
		data = resized
		mimeType = imageformat.DetectMimeType(data)
	}

	if format != "" && format != imageformat.MimeToFormat(mimeType) {
		converted, newMime, err := resizer.Convert(data, format)
		if err != nil {
			return nil, "", fmt.Errorf("convert placeholder: %w", err)
		}
		data = converted
		mimeType = newMime
	}

	return data, mimeType, nil
}
