package testutil

import (
	"context"
	"testing"

	"avatar/internal/domain/model"
	"avatar/internal/imageformat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryAvatarStoreCRUD(t *testing.T) {
	repo := NewMemoryAvatarStore()
	ctx := context.Background()

	avatar := &model.Avatar{
		UserID:           "user@test.com",
		FileName:         "a.jpg",
		MimeType:         imageformat.MIMEJPEG,
		SizeBytes:        10,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(ctx, avatar))
	assert.NotEmpty(t, avatar.ID)

	got, err := repo.GetByID(ctx, avatar.ID)
	require.NoError(t, err)
	assert.Equal(t, avatar.UserID, got.UserID)

	list, err := repo.ListByUserID(ctx, avatar.UserID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	latest, err := repo.GetLatestByUserID(ctx, avatar.UserID)
	require.NoError(t, err)
	assert.Equal(t, avatar.ID, latest.ID)

	require.NoError(t, repo.UpdateProcessingStatus(ctx, avatar.ID, model.ProcessingStatusProcessing))
	require.NoError(t, repo.CompleteProcessing(ctx, avatar.ID, 100, 100))
	got, err = repo.GetByID(ctx, avatar.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ProcessingStatusCompleted, got.ProcessingStatus)

	require.NoError(t, repo.SoftDelete(ctx, avatar.ID))
	_, err = repo.GetByID(ctx, avatar.ID)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestMemoryStorage(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	require.NoError(t, store.SaveOriginal(ctx, "avatar-1", []byte("data"), "text/plain"))
	data, err := store.OpenOriginal(ctx, "avatar-1")
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	require.NoError(t, store.DeleteAll(ctx, "avatar-1"))
	_, err = store.OpenOriginal(ctx, "avatar-1")
	require.Error(t, err)
	require.NoError(t, store.Ping(ctx))
}

func TestNoopPublisher(t *testing.T) {
	p := NoopPublisher{}
	require.NoError(t, p.PublishUploadEvent(context.Background(), model.AvatarUploadEvent{}))
	require.NoError(t, p.PublishDeleteEvent(context.Background(), model.AvatarDeleteEvent{}))
}
