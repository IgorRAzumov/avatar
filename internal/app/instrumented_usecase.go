package app

import (
	"context"
	"time"

	"avatar/internal/domain/model"
	"avatar/internal/domain/usecase"
	"avatar/internal/observability"
)

type instrumentedReadUsecase struct {
	inner usecase.AvatarReadUsecase
	kit   observability.Kit
}

func newInstrumentedReadUsecase(inner usecase.AvatarReadUsecase, kit observability.Kit) usecase.AvatarReadUsecase {
	return &instrumentedReadUsecase{inner: inner, kit: kit}
}

func (usecase *instrumentedReadUsecase) GetImage(
	ctx context.Context, avatarID, size, format string,
) (*model.AvatarImage, error) {
	return observability.RunResult(usecase.kit, ctx, "get_avatar_image", func(ctx context.Context) (*model.AvatarImage, error) {
		return usecase.inner.GetImage(ctx, avatarID, size, format)
	},
		observability.String("avatar_id", avatarID),
		observability.String("size", size),
		observability.String("format", format),
	)
}

func (usecase *instrumentedReadUsecase) GetByID(ctx context.Context, avatarID string) (*model.Avatar, error) {
	return observability.RunResult(usecase.kit, ctx, "get_avatar_metadata", func(ctx context.Context) (*model.Avatar, error) {
		return usecase.inner.GetByID(ctx, avatarID)
	}, observability.String("avatar_id", avatarID))
}

func (usecase *instrumentedReadUsecase) ListByUser(ctx context.Context, userID string) ([]*model.Avatar, error) {
	return observability.RunResult(usecase.kit, ctx, "list_user_avatars", func(ctx context.Context) ([]*model.Avatar, error) {
		return usecase.inner.ListByUser(ctx, userID)
	}, observability.String("user_id", userID))
}

func (usecase *instrumentedReadUsecase) GetLatestByUser(ctx context.Context, userID string) (*model.Avatar, error) {
	return observability.RunResult(usecase.kit, ctx, "get_latest_user_avatar", func(ctx context.Context) (*model.Avatar, error) {
		return usecase.inner.GetLatestByUser(ctx, userID)
	}, observability.String("user_id", userID))
}

func (usecase *instrumentedReadUsecase) GetUserAvatarImage(
	ctx context.Context, userID, size, format string,
) (*model.AvatarImage, error) {
	return observability.RunResult(usecase.kit, ctx, "get_user_avatar_image", func(ctx context.Context) (*model.AvatarImage, error) {
		return usecase.inner.GetUserAvatarImage(ctx, userID, size, format)
	},
		observability.String("user_id", userID),
		observability.String("size", size),
		observability.String("format", format),
	)
}

type instrumentedWriteUsecase struct {
	inner usecase.AvatarWriteUsecase
	read  usecase.AvatarReadUsecase
	kit   observability.Kit
}

func newInstrumentedWriteUsecase(
	inner usecase.AvatarWriteUsecase,
	read usecase.AvatarReadUsecase,
	kit observability.Kit,
) usecase.AvatarWriteUsecase {
	return &instrumentedWriteUsecase{inner: inner, read: read, kit: kit}
}

func (usecase *instrumentedWriteUsecase) Upload(
	ctx context.Context, userID, fileName string, data []byte,
) (*model.Avatar, error) {
	start := time.Now()
	status := "success"
	defer func() {
		usecase.kit.Metrics().RecordUpload(ctx, status, time.Since(start))
	}()

	avatar, err := observability.RunResult(usecase.kit, ctx, "upload_avatar", func(ctx context.Context) (*model.Avatar, error) {
		avatar, uploadErr := usecase.inner.Upload(ctx, userID, fileName, data)
		if uploadErr != nil {
			return nil, uploadErr
		}

		usecase.kit.Metrics().AddStorageBytes(ctx, avatar.SizeBytes)
		usecase.kit.SetAttributes(ctx, observability.String("avatar_id", avatar.ID))
		return avatar, nil
	},
		observability.String("user_id", userID),
		observability.String("file_name", fileName),
		observability.Int64("file_size", int64(len(data))),
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
		usecase.kit.Metrics().RecordDelete(ctx, status, time.Since(start))
	}()

	var sizeBytes int64
	if avatar, err := usecase.read.GetByID(ctx, avatarID); err == nil && avatar != nil {
		sizeBytes = avatar.SizeBytes
	}

	err := usecase.kit.Run(ctx, "delete_avatar", func(ctx context.Context) error {
		return usecase.inner.Delete(ctx, avatarID, requestUserID)
	},
		observability.String("avatar_id", avatarID),
		observability.String("user_id", requestUserID),
	)
	if err != nil {
		status = "error"
		return err
	}

	usecase.kit.Metrics().SubStorageBytes(ctx, sizeBytes)
	return nil
}

func (usecase *instrumentedWriteUsecase) DeleteByUser(ctx context.Context, userID, requestUserID string) error {
	start := time.Now()
	status := "success"
	defer func() {
		usecase.kit.Metrics().RecordDelete(ctx, status, time.Since(start))
	}()

	var sizeBytes int64
	if avatar, err := usecase.read.GetLatestByUser(ctx, userID); err == nil && avatar != nil {
		sizeBytes = avatar.SizeBytes
	}

	err := usecase.kit.Run(ctx, "delete_user_avatar", func(ctx context.Context) error {
		return usecase.inner.DeleteByUser(ctx, userID, requestUserID)
	}, observability.String("user_id", userID))
	if err != nil {
		status = "error"
		return err
	}

	usecase.kit.Metrics().SubStorageBytes(ctx, sizeBytes)
	return nil
}
