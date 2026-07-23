package write_test

import (
	"context"
	"testing"

	"avatar/internal/domain/model"
	"avatar/internal/domain/usecase/write"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvatarCommandUploadValidation(t *testing.T) {
	_, command, _, _ := testutil.NewAvatarUseCases()

	_, err := command.Upload(context.Background(), "", "photo.jpg", testutil.JPEG(10, 10))
	assert.ErrorIs(t, err, model.ErrMissingUserID)

	_, err = command.Upload(context.Background(), "user@test.com", "photo.txt", []byte("plain text"))
	assert.ErrorIs(t, err, model.ErrInvalidFormat)

	cmdSmall := write.NewWriteUsecase(testutil.NewMemoryAvatarStore(), testutil.NewMemoryAvatarStore(), testutil.NewMemoryStorage(), testutil.NoopPublisher{}, 16)
	_, err = cmdSmall.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(100, 100))
	assert.ErrorIs(t, err, model.ErrFileTooLarge)
}

func TestAvatarCommandDeleteForbidden(t *testing.T) {
	_, command, repo, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", testutil.JPEG(20, 20))
	require.NoError(t, err)

	err = command.Delete(context.Background(), resp.ID, "other@test.com")
	assert.ErrorIs(t, err, model.ErrForbidden)
	assert.Contains(t, repo.Avatars, resp.ID)
}

func TestAvatarCommandDeleteSuccess(t *testing.T) {
	_, command, repo, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", testutil.JPEG(20, 20))
	require.NoError(t, err)

	err = command.Delete(context.Background(), resp.ID, "owner@test.com")
	require.NoError(t, err)
	assert.NotContains(t, repo.Avatars, resp.ID)
}

func TestAvatarCommandDeleteByUser(t *testing.T) {
	_, command, repo, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "owner@test.com", "photo.jpg", testutil.JPEG(10, 10))
	require.NoError(t, err)

	err = command.DeleteByUser(context.Background(), "owner@test.com", "owner@test.com")
	require.NoError(t, err)
	assert.NotContains(t, repo.Avatars, resp.ID)

	err = command.DeleteByUser(context.Background(), "owner@test.com", "other@test.com")
	assert.ErrorIs(t, err, model.ErrForbidden)
}

func TestAvatarCommandUpload(t *testing.T) {
	_, command, _, _ := testutil.NewAvatarUseCases()
	data := testutil.JPEG(640, 480)

	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", data)
	require.NoError(t, err)
	assert.Equal(t, "user@test.com", resp.UserID)
	assert.Equal(t, model.ProcessingStatusPending, resp.ProcessingStatus)
	assert.NotEmpty(t, resp.ID)
}
