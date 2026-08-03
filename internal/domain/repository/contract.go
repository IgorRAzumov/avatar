package repository

import (
	"context"

	"avatar/internal/domain/model"
)

type ReadRepository interface {
	GetByID(ctx context.Context, id string) (*model.Avatar, error)
	GetLatestByUserID(ctx context.Context, userID string) (*model.Avatar, error)
	ListByUserID(ctx context.Context, userID string) ([]*model.Avatar, error)
}

type WriteRepository interface {
	Create(ctx context.Context, avatar *model.Avatar) error
	SoftDelete(ctx context.Context, id string) error
	UpdateProcessingStatus(ctx context.Context, id, status string) error
	CompleteProcessing(ctx context.Context, id string, width, height int) error
}

type FilesRepository interface {
	SaveOriginal(ctx context.Context, avatarID string, data []byte, contentType string) error
	SaveThumbnail(ctx context.Context, avatarID, size string, data []byte) error
	OpenOriginal(ctx context.Context, avatarID string) ([]byte, error)
	OpenThumbnail(ctx context.Context, avatarID, size string) ([]byte, error)
	DeleteAll(ctx context.Context, avatarID string) error
}

type EventPublisherRepository interface {
	PublishUploadEvent(ctx context.Context, event model.AvatarUploadEvent) error
	PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error
}

type HealthRepository interface {
	Check(ctx context.Context) model.HealthReport
}
