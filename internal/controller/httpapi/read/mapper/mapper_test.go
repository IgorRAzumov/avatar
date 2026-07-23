package mapper_test

import (
	"testing"
	"time"

	"avatar/internal/controller/httpapi/read/mapper"
	domainmodel "avatar/internal/domain/model"

	"github.com/stretchr/testify/assert"
)

func TestAvatarMetadataResponse(t *testing.T) {
	now := time.Now().UTC()
	avatar := &domainmodel.Avatar{
		ID:        "avatar-1",
		UserID:    "user@test.com",
		FileName:  "photo.jpg",
		MimeType:  "image/jpeg",
		SizeBytes: 1024,
		Width:     640,
		Height:    480,
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := mapper.AvatarMetadataResponse("http://host", avatar.ID, avatar)

	assert.Equal(t, "avatar-1", resp.ID)
	assert.Equal(t, "user@test.com", resp.UserID)
	assert.Equal(t, "photo.jpg", resp.FileName)
	assert.Equal(t, int64(1024), resp.Size)
	assert.Equal(t, 640, resp.Dimensions.Width)
	assert.Equal(t, 480, resp.Dimensions.Height)
	assert.Len(t, resp.Thumbnails, len(domainmodel.ThumbnailSizes))
	for _, thumb := range resp.Thumbnails {
		assert.NotEmpty(t, thumb.Size)
		assert.Contains(t, thumb.URL, "http://host")
	}
}
