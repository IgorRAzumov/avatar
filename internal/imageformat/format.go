package imageformat

import (
	"net/http"
	"path/filepath"
	"strings"
)

const (
	JPEG = "jpeg"
	PNG  = "png"
	WebP = "webp"

	MIMEJPEG = "image/jpeg"
	MIMEPNG  = "image/png"
	MIMEWebP = "image/webp"
)

const (
	JPEGQuality = 85
	DefaultMIME = MIMEJPEG
)

var Supported = []string{JPEG, PNG, WebP}

var mimeToFormat = map[string]string{
	MIMEJPEG: JPEG,
	MIMEPNG:  PNG,
	MIMEWebP: WebP,
}

var extToMIME = map[string]string{
	".jpg":  MIMEJPEG,
	".jpeg": MIMEJPEG,
	".png":  MIMEPNG,
	".webp": MIMEWebP,
}

func IsSupportedMimeType(mimeType string) bool {
	_, ok := mimeToFormat[mimeType]
	return ok
}

func MimeToFormat(mimeType string) string {
	return mimeToFormat[mimeType]
}

func FormatToMime(format string) string {
	switch NormalizeFormat(format) {
	case JPEG:
		return MIMEJPEG
	case PNG:
		return MIMEPNG
	case WebP:
		return MIMEWebP
	default:
		return DefaultMIME
	}
}

func NormalizeFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpg", JPEG:
		return JPEG
	case PNG:
		return PNG
	case WebP:
		return WebP
	default:
		return format
	}
}

func MIMEFromFileName(name string) string {
	return extToMIME[strings.ToLower(filepath.Ext(name))]
}

func SupportedFormatsLabel() string {
	return strings.Join(Supported, ", ")
}

// DetectMimeType sniffs the content type from the file header using the
// standard library and returns one of the supported image MIME types, or an
// empty string if the data is not a recognised supported image.
func DetectMimeType(data []byte) string {
	switch http.DetectContentType(data) {
	case MIMEJPEG:
		return MIMEJPEG
	case MIMEPNG:
		return MIMEPNG
	case MIMEWebP:
		return MIMEWebP
	default:
		return ""
	}
}
