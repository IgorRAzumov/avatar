package usecase

import (
	"context"

	"avatar/internal/domain/model"
)

type AvatarReadUsecase interface {
	GetImage(ctx context.Context, avatarID, size, format string) (*model.AvatarImage, error)
	GetByID(ctx context.Context, avatarID string) (*model.Avatar, error)
	ListByUser(ctx context.Context, userID string) ([]*model.Avatar, error)
	GetLatestByUser(ctx context.Context, userID string) (*model.Avatar, error)
	GetUserAvatarImage(ctx context.Context, userID, size, format string) (*model.AvatarImage, error)
}

type AvatarWriteUsecase interface {
	Upload(ctx context.Context, userID, fileName string, data []byte) (*model.Avatar, error)
	Delete(ctx context.Context, avatarID, requestUserID string) error
	DeleteByUser(ctx context.Context, userID, requestUserID string) error
}
