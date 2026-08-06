package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"avatar/internal/domain/model"
	"avatar/internal/imageformat"
	"avatar/internal/logger"
	"avatar/internal/observability"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageProcessorProcessUpload(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	store := testutil.NewMemoryStorage()
	proc := NewImageProcessor(repo, repo, store, observability.NewTestKit())

	data := testutil.JPEG(300, 200)

	avatar := &model.Avatar{
		UserID:           "user@test.com",
		FileName:         "photo.jpg",
		MimeType:         imageformat.MIMEJPEG,
		SizeBytes:        int64(len(data)),
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(context.Background(), avatar))
	require.NoError(t, store.SaveOriginal(context.Background(), avatar.ID, data, imageformat.MIMEJPEG))

	err := proc.ProcessUpload(context.Background(), model.AvatarUploadEvent{
		AvatarID: avatar.ID,
		UserID:   avatar.UserID,
	})
	require.NoError(t, err)
	assert.Equal(t, model.ProcessingStatusCompleted, repo.Avatars[avatar.ID].ProcessingStatus)

	thumb, err := store.OpenThumbnail(context.Background(), avatar.ID, model.ImageSizeSmall)
	require.NoError(t, err)
	assert.NotEmpty(t, thumb)
}

func TestImageProcessorSkipsCompletedAvatar(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	proc := NewImageProcessor(repo, repo, testutil.NewMemoryStorage(), observability.NewTestKit())

	avatar := &model.Avatar{
		UserID:           "user@test.com",
		ProcessingStatus: model.ProcessingStatusCompleted,
	}
	require.NoError(t, repo.Create(context.Background(), avatar))

	err := proc.ProcessUpload(context.Background(), model.AvatarUploadEvent{
		AvatarID: avatar.ID,
	})
	require.NoError(t, err)
}

func TestAsyncPublisherPublishDeleteEvent(t *testing.T) {
	store := testutil.NewMemoryStorage()
	require.NoError(t, store.SaveOriginal(context.Background(), "avatar-1", []byte("a"), "text/plain"))
	proc := NewImageProcessor(nil, nil, store, observability.NewTestKit())
	p := NewAsyncPublisher(proc, logger.Nop(), observability.NewTestKit())

	require.NoError(t, p.PublishDeleteEvent(context.Background(), model.AvatarDeleteEvent{
		AvatarID: "avatar-1",
	}))

	require.Eventually(t, func() bool {
		_, err := store.OpenOriginal(context.Background(), "avatar-1")
		return errors.Is(err, model.ErrNotFound)
	}, time.Second, 5*time.Millisecond)
}

func TestAsyncPublisherPublishUploadEventSurvivesParentCancellation(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	store := testutil.NewMemoryStorage()
	p := NewAsyncPublisher(NewImageProcessor(repo, repo, store, observability.NewTestKit()), logger.Nop(), observability.NewTestKit())

	data := testutil.JPEG(120, 90)
	avatar := &model.Avatar{
		ID:               "avatar-1",
		UserID:           "user@test.com",
		FileName:         "photo.jpg",
		MimeType:         imageformat.MIMEJPEG,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(context.Background(), avatar))
	require.NoError(t, store.SaveOriginal(context.Background(), avatar.ID, data, imageformat.MIMEJPEG))

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, p.PublishUploadEvent(ctx, model.AvatarUploadEvent{
		AvatarID: avatar.ID,
		UserID:   avatar.UserID,
	}))
	cancel() // handler returned / request context cancelled

	require.Eventually(t, func() bool {
		return repo.ProcessingStatus(avatar.ID) == model.ProcessingStatusCompleted
	}, time.Second, 5*time.Millisecond)
}

func TestAsyncPublisherPublishUploadEvent(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	store := testutil.NewMemoryStorage()
	p := NewAsyncPublisher(NewImageProcessor(repo, repo, store, observability.NewTestKit()), logger.Nop(), observability.NewTestKit())

	data := testutil.JPEG(120, 90)
	avatar := &model.Avatar{
		UserID:           "user@test.com",
		FileName:         "photo.jpg",
		MimeType:         imageformat.MIMEJPEG,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(context.Background(), avatar))
	require.NoError(t, store.SaveOriginal(context.Background(), avatar.ID, data, imageformat.MIMEJPEG))

	require.NoError(t, p.PublishUploadEvent(context.Background(), model.AvatarUploadEvent{
		AvatarID: avatar.ID,
		UserID:   avatar.UserID,
	}))

	require.Eventually(t, func() bool {
		return repo.ProcessingStatus(avatar.ID) == model.ProcessingStatusCompleted
	}, time.Second, 5*time.Millisecond)
}
