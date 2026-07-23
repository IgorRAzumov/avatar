package processor

import (
	"context"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/logger"
	"avatar/internal/retry"
)

type AsyncPublisher struct {
	processor *ImageProcessor
	logger    *logger.Logger
}

func NewAsyncPublisher(processor *ImageProcessor, log *logger.Logger) *AsyncPublisher {
	return &AsyncPublisher{processor: processor, logger: log}
}

func NewAsyncImageResizerProcessor(
	readRepository repository.ReadRepository,
	writeRepository repository.WriteRepository,
	files repository.FilesRepository,
	logger *logger.Logger,
) *AsyncPublisher {
	return NewAsyncPublisher(NewImageProcessor(readRepository, writeRepository, files), logger)
}

func backgroundContext(parent context.Context) context.Context {
	return context.WithoutCancel(parent)
}

func (publisher *AsyncPublisher) PublishUploadEvent(ctx context.Context, event model.AvatarUploadEvent) error {
	go func(ctx context.Context) {
		defer publisher.recoverPanic("process upload", event.AvatarID)

		err := retry.WithBackoff(ctx, retry.DefaultMaxAttempts,
			func() error {
				return publisher.processor.ProcessUpload(ctx, event)
			},
			func(attempt int, err error) {
				publisher.logger.Warn("thumbnail processing failed", "avatar_id", event.AvatarID, "attempt", attempt, "error", err)
			},
		)
		if err != nil {
			publisher.logger.Error("thumbnail processing failed after retries", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error {
	go func(ctx context.Context) {
		defer publisher.recoverPanic("process delete", event.AvatarID)
		if err := publisher.processor.ProcessDelete(ctx, event); err != nil {
			publisher.logger.Error("delete avatar files failed", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) recoverPanic(op, avatarID string) {
	if r := recover(); r != nil {
		publisher.logger.Error("panic in async processing", "op", op, "avatar_id", avatarID, "panic", r)
	}
}
