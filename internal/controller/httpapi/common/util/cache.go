package util

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// ImageETag returns a strong ETag for binary image data (quoted hex SHA-256).
func ImageETag(data []byte) string {
	sum := sha256.Sum256(data)
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

// WriteCachedImage writes image bytes with Cache-Control and ETag headers.
// When If-None-Match matches the computed ETag, responds with 304 Not Modified.
func WriteCachedImage(writer http.ResponseWriter, request *http.Request, data []byte, mimeType, cacheControl string) {
	etag := ImageETag(data)
	writer.Header().Set("ETag", etag)
	writer.Header().Set("Cache-Control", cacheControl)
	writer.Header().Set("Content-Type", mimeType)

	if request.Header.Get("If-None-Match") == etag {
		writer.WriteHeader(http.StatusNotModified)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(data)
}
