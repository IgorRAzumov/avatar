package process

import (
	"context"
	"time"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/imageutil"
	"avatar/internal/observability"
)

type ImageProcessor struct {
	readRepository  repository.ReadRepository
	writeRepository repository.WriteRepository
	filesRepository repository.FilesRepository
	resizer         *imageutil.Resizer
	kit             observability.Kit
}

func NewImageProcessor(
	readRepository repository.ReadRepository,
	writeRepository repository.WriteRepository,
	files repository.FilesRepository,
	kit observability.Kit,
) *ImageProcessor {
	return &ImageProcessor{
		readRepository:  readRepository,
		writeRepository: writeRepository,
		filesRepository: files,
		resizer:         imageutil.NewResizer(),
		kit:             kit,
	}
}

func (processor *ImageProcessor) ProcessUpload(ctx context.Context, event model.AvatarUploadEvent) error {
	start := time.Now()
	status := "success"
	defer func() {
		processor.kit.Metrics().RecordProcessing(ctx, "upload", status, time.Since(start))
	}()

	err := processor.kit.Run(ctx, "process_upload", func(ctx context.Context) error {
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
		processor.kit.SetAttributes(ctx,
			observability.Int("image.width", width),
			observability.Int("image.height", height),
		)

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
	},
		observability.String("avatar_id", event.AvatarID),
		observability.String("user_id", event.UserID),
	)
	if err != nil {
		status = "error"
	}
	return err
}

func (processor *ImageProcessor) ProcessDelete(ctx context.Context, event model.AvatarDeleteEvent) error {
	start := time.Now()
	status := "success"
	defer func() {
		processor.kit.Metrics().RecordProcessing(ctx, "delete", status, time.Since(start))
	}()

	err := processor.kit.Run(ctx, "process_delete", func(ctx context.Context) error {
		return processor.filesRepository.DeleteAll(ctx, event.AvatarID)
	}, observability.String("avatar_id", event.AvatarID))
	if err != nil {
		status = "error"
	}
	return err
}
