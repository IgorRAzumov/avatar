package mapper

import (
	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/controller/httpapi/read/model"
	domainmodel "avatar/internal/domain/model"
)

func AvatarMetadataResponse(baseURL, avatarID string, avatar *domainmodel.Avatar) model.AvatarMetadataResponse {
	thumbnails := make([]model.ThumbnailInfo, 0, len(domainmodel.ThumbnailSizes))
	for _, thumbnail := range domainmodel.ThumbnailSizes {
		thumbnails = append(thumbnails, model.ThumbnailInfo{
			Size: thumbnail.Name,
			URL:  util.AvatarURLWithParams(baseURL, avatarID, thumbnail.Name, ""),
		})
	}

	return model.AvatarMetadataResponse{
		ID:       avatar.ID,
		UserID:   avatar.UserID,
		FileName: avatar.FileName,
		MimeType: avatar.MimeType,
		Size:     avatar.SizeBytes,
		Dimensions: model.Dimensions{
			Width:  avatar.Width,
			Height: avatar.Height,
		},
		Thumbnails: thumbnails,
		CreatedAt:  avatar.CreatedAt,
		UpdatedAt:  avatar.UpdatedAt,
	}
}
