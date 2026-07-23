package processor

import (
	"context"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/imageutil"
)

type ImageProcessor struct {
	readRepository  repository.ReadRepository
	writeRepository repository.WriteRepository
	filesRepository repository.FilesRepository
	resizer         *imageutil.Resizer
}

func NewImageProcessor(
	readRepository repository.ReadRepository,
	writeRepository repository.WriteRepository,
	files repository.FilesRepository,
) *ImageProcessor {
	return &ImageProcessor{
		readRepository:  readRepository,
		writeRepository: writeRepository,
		filesRepository: files,
		resizer:         imageutil.NewResizer(),
	}
}

func (processor *ImageProcessor) ProcessUpload(ctx context.Context, event model.AvatarUploadEvent) error {
	avatar, err := processor.readRepository.GetByID(ctx, event.AvatarID)
	if err != nil {
		return err
	}
	if avatar.ProcessingStatus == model.ProcessingStatusCompleted {
		return nil
	}

	if err := processor.writeRepository.UpdateProcessingStatus(
		ctx, event.AvatarID, model.ProcessingStatusProcessing,
	); err != nil {
		return err
	}

	image, err := processor.filesRepository.OpenOriginal(ctx, event.AvatarID)
	if err != nil {
		_ = processor.writeRepository.UpdateProcessingStatus(ctx, event.AvatarID, model.ProcessingStatusFailed)
		return err
	}

	width, height, _ := processor.resizer.Dimensions(image)

	for _, size := range model.ThumbnailSizes {
		data, resizeErr := processor.resizer.ResizeAsJPEG(image, size.Width, size.Height)
		if resizeErr != nil {
			_ = processor.writeRepository.UpdateProcessingStatus(ctx, event.AvatarID, model.ProcessingStatusFailed)
			return resizeErr
		}
		if err := processor.filesRepository.SaveThumbnail(ctx, event.AvatarID, size.Name, data); err != nil {
			_ = processor.writeRepository.UpdateProcessingStatus(ctx, event.AvatarID, model.ProcessingStatusFailed)
			return err
		}
	}

	return processor.writeRepository.CompleteProcessing(ctx, event.AvatarID, width, height)
}

func (processor *ImageProcessor) ProcessDelete(ctx context.Context, event model.AvatarDeleteEvent) error {
	return processor.filesRepository.DeleteAll(ctx, event.AvatarID)
}
