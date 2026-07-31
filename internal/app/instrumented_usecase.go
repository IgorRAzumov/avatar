package app

import (
	"context"
	"time"

	"avatar/internal/domain/model"
	"avatar/internal/domain/usecase"
	"avatar/internal/observability"

	"go.opentelemetry.io/otel/attribute"
)

type instrumentedReadUsecase struct {
	inner usecase.AvatarReadUsecase
}

func newInstrumentedReadUsecase(inner usecase.AvatarReadUsecase) usecase.AvatarReadUsecase {
	return &instrumentedReadUsecase{inner: inner}
}

func (usecase *instrumentedReadUsecase) GetImage(
	ctx context.Context, avatarID, size, format string,
) (*model.AvatarImage, error) {
	return observability.RunResult(ctx, "get_avatar_image", func(ctx context.Context) (*model.AvatarImage, error) {
		return usecase.inner.GetImage(ctx, avatarID, size, format)
	},
		attribute.String("avatar_id", avatarID),
		attribute.String("size", size),
		attribute.String("format", format),
	)
}

func (usecase *instrumentedReadUsecase) GetByID(ctx context.Context, avatarID string) (*model.Avatar, error) {
	return observability.RunResult(ctx, "get_avatar_metadata", func(ctx context.Context) (*model.Avatar, error) {
		return usecase.inner.GetByID(ctx, avatarID)
	}, attribute.String("avatar_id", avatarID))
}

func (usecase *instrumentedReadUsecase) ListByUser(ctx context.Context, userID string) ([]*model.Avatar, error) {
	return observability.RunResult(ctx, "list_user_avatars", func(ctx context.Context) ([]*model.Avatar, error) {
		return usecase.inner.ListByUser(ctx, userID)
	}, attribute.String("user_id", userID))
}

func (usecase *instrumentedReadUsecase) GetUserAvatarImage(
	ctx context.Context, userID, size, format string,
) (*model.AvatarImage, error) {
	return observability.RunResult(ctx, "get_user_avatar_image", func(ctx context.Context) (*model.AvatarImage, error) {
		return usecase.inner.GetUserAvatarImage(ctx, userID, size, format)
	},
		attribute.String("user_id", userID),
		attribute.String("size", size),
		attribute.String("format", format),
	)
}

type instrumentedWriteUsecase struct {
	inner usecase.AvatarWriteUsecase
	read  usecase.AvatarReadUsecase
}

func newInstrumentedWriteUsecase(
	inner usecase.AvatarWriteUsecase,
	read usecase.AvatarReadUsecase,
) usecase.AvatarWriteUsecase {
	return &instrumentedWriteUsecase{inner: inner, read: read}
}

func (usecase *instrumentedWriteUsecase) Upload(
	ctx context.Context, userID, fileName string, data []byte,
) (*model.Avatar, error) {
	start := time.Now()
	status := "success"
	defer func() {
		observability.RecordUpload(status, time.Since(start))
	}()

	avatar, err := observability.RunResult(ctx, "upload_avatar", func(ctx context.Context) (*model.Avatar, error) {
		avatar, uploadErr := usecase.inner.Upload(ctx, userID, fileName, data)
		if uploadErr != nil {
			return nil, uploadErr
		}

		observability.AddStorageBytes(avatar.SizeBytes)
		observability.SetAttributes(ctx, attribute.String("avatar_id", avatar.ID))
		return avatar, nil
	},
		attribute.String("user_id", userID),
		attribute.String("file_name", fileName),
		attribute.Int64("file_size", int64(len(data))),
	)
	if err != nil {
		status = "error"
	}
	return avatar, err
}

func (usecase *instrumentedWriteUsecase) Delete(ctx context.Context, avatarID, requestUserID string) error {
	start := time.Now()
	status := "success"
	defer func() {
		observability.RecordDelete(status, time.Since(start))
	}()

	var sizeBytes int64
	if avatar, err := usecase.read.GetByID(ctx, avatarID); err == nil && avatar != nil {
		sizeBytes = avatar.SizeBytes
	}

	err := observability.Run(ctx, "delete_avatar", func(ctx context.Context) error {
		return usecase.inner.Delete(ctx, avatarID, requestUserID)
	},
		attribute.String("avatar_id", avatarID),
		attribute.String("user_id", requestUserID),
	)
	if err != nil {
		status = "error"
		return err
	}

	observability.SubStorageBytes(sizeBytes)
	return nil
}

func (usecase *instrumentedWriteUsecase) DeleteByUser(ctx context.Context, userID, requestUserID string) error {
	start := time.Now()
	status := "success"
	defer func() {
		observability.RecordDelete(status, time.Since(start))
	}()

	var sizeBytes int64
	if avatars, err := usecase.read.ListByUser(ctx, userID); err == nil && len(avatars) > 0 {
		sizeBytes = avatars[0].SizeBytes
	}

	err := observability.Run(ctx, "delete_user_avatar", func(ctx context.Context) error {
		return usecase.inner.DeleteByUser(ctx, userID, requestUserID)
	}, attribute.String("user_id", userID))
	if err != nil {
		status = "error"
		return err
	}

	observability.SubStorageBytes(sizeBytes)
	return nil
}
