package mapper

import (
	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/controller/httpapi/write/model"
	domainmodel "avatar/internal/domain/model"
)

func UploadResponse(baseURL string, avatar *domainmodel.Avatar) model.UploadResponse {
	return model.UploadResponse{
		ID:        avatar.ID,
		UserID:    avatar.UserID,
		URL:       util.AvatarURL(baseURL, avatar.ID),
		Status:    domainmodel.ProcessingStatusProcessing,
		CreatedAt: avatar.CreatedAt,
	}
}
