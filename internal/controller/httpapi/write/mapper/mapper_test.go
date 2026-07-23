package mapper_test

import (
	"testing"
	"time"

	"avatar/internal/controller/httpapi/write/mapper"
	domainmodel "avatar/internal/domain/model"

	"github.com/stretchr/testify/assert"
)

func TestUploadResponse(t *testing.T) {
	now := time.Now().UTC()
	avatar := &domainmodel.Avatar{
		ID:        "avatar-1",
		UserID:    "user@test.com",
		CreatedAt: now,
	}

	resp := mapper.UploadResponse("http://host/", avatar)

	assert.Equal(t, "avatar-1", resp.ID)
	assert.Equal(t, "user@test.com", resp.UserID)
	assert.Equal(t, domainmodel.ProcessingStatusProcessing, resp.Status)
	assert.Equal(t, "http://host/api/v1/avatars/avatar-1", resp.URL)
	assert.Equal(t, now, resp.CreatedAt)
}
