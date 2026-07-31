package processor

import (
	"context"

	"avatar/internal/domain/model"
	"avatar/internal/domain/repository"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/retry"

	"go.opentelemetry.io/otel/attribute"
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
		log := publisher.logger.WithContext(ctx)
		defer publisher.recoverPanic(ctx, "process upload", event.AvatarID)

		err := observability.Run(ctx, "async.process_upload", func(ctx context.Context) error {
			return retry.WithBackoff(ctx, retry.DefaultMaxAttempts,
				func() error {
					return publisher.processor.ProcessUpload(ctx, event)
				},
				func(attempt int, err error) {
					observability.AddEvent(ctx, "retry",
						attribute.Int("attempt", attempt),
						attribute.String("error", err.Error()),
					)
					log.Warn("thumbnail processing failed", "avatar_id", event.AvatarID, "attempt", attempt, "error", err)
				},
			)
		},
			attribute.String("avatar_id", event.AvatarID),
			attribute.String("user_id", event.UserID),
		)
		if err != nil {
			log.Error("thumbnail processing failed after retries", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) PublishDeleteEvent(ctx context.Context, event model.AvatarDeleteEvent) error {
	go func(ctx context.Context) {
		log := publisher.logger.WithContext(ctx)
		defer publisher.recoverPanic(ctx, "process delete", event.AvatarID)

		err := observability.Run(ctx, "async.process_delete", func(ctx context.Context) error {
			return publisher.processor.ProcessDelete(ctx, event)
		}, attribute.String("avatar_id", event.AvatarID))
		if err != nil {
			log.Error("delete avatar files failed", "avatar_id", event.AvatarID, "error", err)
		}
	}(backgroundContext(ctx))
	return nil
}

func (publisher *AsyncPublisher) recoverPanic(ctx context.Context, op, avatarID string) {
	if r := recover(); r != nil {
		publisher.logger.WithContext(ctx).Error("panic in async processing", "op", op, "avatar_id", avatarID, "panic", r)
	}
}
