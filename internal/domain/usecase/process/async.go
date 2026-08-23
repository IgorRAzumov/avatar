package process

import (
	"context"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/retry"
)

type AsyncPublisher struct {
	processor *ImageProcessor
	logger    *logger.Logger
	kit       observability.Kit
}

func NewAsyncPublisher(processor *ImageProcessor, log *logger.Logger, kit observability.Kit) *AsyncPublisher {
	return &AsyncPublisher{processor: processor, logger: log, kit: kit}
}

func NewAsyncImageResizerProcessor(
	readRepository repository.ReadRepository,
	writeRepository repository.WriteRepository,
	files repository.FilesRepository,
	log *logger.Logger,
	kit observability.Kit,
) *AsyncPublisher {
	return NewAsyncPublisher(NewImageProcessor(readRepository, writeRepository, files, kit), log, kit)
}

func backgroundContext(parent context.Context) context.Context {
	return context.WithoutCancel(parent)
}

func (publisher *AsyncPublisher) PublishUploadEvent(ctx context.Context, event model.AvatarUploadEvent) error {
	go func(ctx context.Context) {
		defer publisher.recoverPanic(ctx, "process upload", event.AvatarID)

		err := publisher.kit.Run(ctx, "async.process_upload", func(ctx context.Context) error {
			return retry.WithBackoff(ctx, retry.DefaultMaxAttempts,
				func() error {
					return publisher.processor.ProcessUpload(ctx, event)
				},
				func(attempt int, err error) {
					publisher.kit.AddEvent(ctx, "retry",
						observability.Int("attempt", attempt),
						observability.String("error", err.Error()),
					)
					publisher.logger.Warn(ctx, "thumbnail processing failed", "avatar_id", event.AvatarID, "attempt", attempt, "error", err)
				},
			)
		},
			observability.String("avatar_id", event.AvatarID),
			observability.String("user_id", event.UserID),
		)
		if err != nil {
			publisher.logger.Error(ctx, "thumbnail processing failed after retries", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error {
	go func(ctx context.Context) {
		defer publisher.recoverPanic(ctx, "process delete", event.AvatarID)

		err := publisher.kit.Run(ctx, "async.process_delete", func(ctx context.Context) error {
			return publisher.processor.ProcessDelete(ctx, event)
		}, observability.String("avatar_id", event.AvatarID))
		if err != nil {
			publisher.logger.Error(ctx, "delete avatar files failed", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) recoverPanic(ctx context.Context, op, avatarID string) {
	if r := recover(); r != nil {
		publisher.logger.Error(ctx, "panic in async processing", "op", op, "avatar_id", avatarID, "panic", r)
	}
}
