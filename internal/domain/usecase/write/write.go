package write

import (
	"context"
	"fmt"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/imageformat"
	"avatar/internal/imageutil"
)

type Usecase struct {
	readRepository  repository.ReadRepository
	writeRepository repository.WriteRepository
	filesRepository repository.FilesRepository
	publisher       repository.EventPublisherRepository
	resizer         *imageutil.Resizer
	maxSize         int64
}

func NewWriteUsecase(
	readRepository repository.ReadRepository,
	writeRepository repository.WriteRepository,
	filesRepository repository.FilesRepository,
	publisherRepository repository.EventPublisherRepository,
	maxSize int64,
) *Usecase {
	return &Usecase{
		readRepository:  readRepository,
		writeRepository: writeRepository,
		filesRepository: filesRepository,
		publisher:       publisherRepository,
		resizer:         imageutil.NewResizer(),
		maxSize:         maxSize,
	}
}

func (usecase *Usecase) Upload(
	ctx context.Context, userID, fileName string, data []byte,
) (*model.Avatar, error) {
	if userID == "" {
		return nil, model.ErrMissingUserID
	}
	if int64(len(data)) > usecase.maxSize {
		return nil, model.ErrFileTooLarge
	}

	mimeType := imageutil.DetectMimeType(data)
	if mimeType == "" {
		if ct := imageformat.MIMEFromFileName(fileName); ct != "" {
			mimeType = ct
		}
	}
	if !imageformat.IsSupportedMimeType(mimeType) {
		return nil, model.ErrInvalidFormat
	}

	width, height, _ := usecase.resizer.Dimensions(data)

	avatar := &model.Avatar{
		UserID:           userID,
		FileName:         fileName,
		MimeType:         mimeType,
		SizeBytes:        int64(len(data)),
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
		Width:            width,
		Height:           height,
	}

	if err := usecase.writeRepository.Create(ctx, avatar); err != nil {
		return nil, fmt.Errorf("create avatar record: %w", err)
	}

	if err := usecase.filesRepository.SaveOriginal(ctx, avatar.ID, data, mimeType); err != nil {
		if delErr := usecase.writeRepository.SoftDelete(ctx, avatar.ID); delErr != nil {
			return nil, fmt.Errorf("upload file: %w (rollback avatar record: %v)", err, delErr)
		}
		return nil, fmt.Errorf("upload file: %w", err)
	}

	event := model.AvatarUploadEvent{
		AvatarID: avatar.ID,
		UserID:   userID,
	}
	if err := usecase.publisher.PublishUploadEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("publish upload event: %w", err)
	}

	return avatar, nil
}

func (usecase *Usecase) Delete(ctx context.Context, avatarID, requestUserID string) error {
	avatar, err := usecase.readRepository.GetByID(ctx, avatarID)
	if err != nil {
		return err
	}
	if avatar.UserID != requestUserID {
		return model.ErrForbidden
	}

	if err := usecase.writeRepository.SoftDelete(ctx, avatarID); err != nil {
		return err
	}

	return usecase.publisher.PublishDeleteEvent(ctx, model.AvatarDeleteEvent{
		AvatarID: avatarID,
	})
}

func (usecase *Usecase) DeleteByUser(ctx context.Context, userID, requestUserID string) error {
	if userID != requestUserID {
		return model.ErrForbidden
	}
	avatar, err := usecase.readRepository.GetLatestByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return usecase.Delete(ctx, avatar.ID, requestUserID)
}
