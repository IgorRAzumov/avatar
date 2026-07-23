// Package objectkey builds the storage keys/paths used to address avatar files
// in any file backend (S3, in-memory). It is backend-agnostic
// so adapters do not need to depend on one another.
package objectkey

import "fmt"

const avatarsPrefix = "avatars"

func OriginalObjectKey(avatarID string) string {
	return fmt.Sprintf("%s/%s/original", avatarsPrefix, avatarID)
}

func ThumbnailObjectKey(avatarID, size string) string {
	return fmt.Sprintf("%s/%s/thumbnails/%s.jpg", avatarsPrefix, avatarID, size)
}

func AvatarDir(avatarID string) string {
	return fmt.Sprintf("%s/%s", avatarsPrefix, avatarID)
}
