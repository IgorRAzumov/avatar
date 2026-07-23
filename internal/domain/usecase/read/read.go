package read

import (
	"context"
	"errors"
	"fmt"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/imageformat"
	"avatar/internal/imageutil"
	"avatar/internal/placeholder"
)

type Usecase struct {
	readRepository repository.ReadRepository
	files          repository.FilesRepository
	resizer        *imageutil.Resizer
}

func NewReadUsecase(
	readRepository repository.ReadRepository,
	filesRepository repository.FilesRepository,
) *Usecase {
	return &Usecase{
		readRepository: readRepository,
		files:          filesRepository,
		resizer:        imageutil.NewResizer(),
	}
}

func (usecase *Usecase) GetImage(
	ctx context.Context, avatarID, size, format string,
) (*model.AvatarImage, error) {
	avatar, err := usecase.readRepository.GetByID(ctx, avatarID)
	if err != nil {
		return nil, err
	}

	var data []byte
	var mimeType string

	switch {
	case model.IsOriginalImageSize(size):
		data, err = usecase.files.OpenOriginal(ctx, avatarID)
		mimeType = avatar.MimeType
	default:
		thumbnail, ok := model.LookupThumbnailSize(size)
		if !ok {
			return nil, model.ErrNotFound
		}
		data, err = usecase.files.OpenThumbnail(ctx, avatarID, thumbnail.Name)
		mimeType = imageformat.MIMEJPEG
	}
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}

	if format != "" && format != imageformat.MimeToFormat(mimeType) {
		converted, newMime, convErr := usecase.resizer.Convert(data, format)
		if convErr != nil {
			return nil, convErr
		}
		data = converted
		mimeType = newMime
	}

	return &model.AvatarImage{
		Data:     data,
		MimeType: mimeType,
	}, nil
}

func (usecase *Usecase) GetByID(ctx context.Context, avatarID string) (*model.Avatar, error) {
	return usecase.readRepository.GetByID(ctx, avatarID)
}

func (usecase *Usecase) ListByUser(ctx context.Context, userID string) ([]*model.Avatar, error) {
	return usecase.readRepository.ListByUserID(ctx, userID)
}

func (usecase *Usecase) GetUserAvatarImage(ctx context.Context, userID, size, format string) (*model.AvatarImage, error) {
	avatar, err := usecase.readRepository.GetLatestByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return usecase.placeholderImage(size, format)
		}
		return nil, err
	}

	return usecase.getUserImageWithFallback(ctx, avatar.ID, size, format)
}

func (usecase *Usecase) getUserImageWithFallback(
	ctx context.Context, avatarID, size, format string,
) (*model.AvatarImage, error) {
	image, err := usecase.GetImage(ctx, avatarID, size, format)
	if err == nil {
		return image, nil
	}
	if !errors.Is(err, model.ErrNotFound) || model.IsOriginalImageSize(size) {
		return nil, err
	}

	return usecase.GetImage(ctx, avatarID, model.ImageSizeOriginal, format)
}

func (usecase *Usecase) placeholderImage(size, format string) (*model.AvatarImage, error) {
	data, mimeType, err := placeholder.Image(size, format)
	if err != nil {
		return nil, err
	}
	return &model.AvatarImage{
		Data:     data,
		MimeType: mimeType,
	}, nil
}
