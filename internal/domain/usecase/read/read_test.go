package read_test

import (
	"context"
	"testing"

	"avatar/internal/domain/model"
	"avatar/internal/domain/usecase/read"
	"avatar/internal/imageformat"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvatarQueryGetImageOriginal(t *testing.T) {
	query, command, _, store := testutil.NewAvatarUseCases()
	data := testutil.JPEG(200, 150)

	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", data)
	require.NoError(t, err)

	image, err := query.GetImage(context.Background(), resp.ID, "", "")
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEJPEG, image.MimeType)
	assert.Equal(t, data, image.Data)
	assert.NotEmpty(t, store.Objects)
}

func TestAvatarQueryGetImageThumbnailMissing(t *testing.T) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(50, 50))
	require.NoError(t, err)

	_, err = query.GetImage(context.Background(), resp.ID, model.ImageSizeSmall, "")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestAvatarQueryListByUser(t *testing.T) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	_, err := command.Upload(context.Background(), "user@test.com", "one.jpg", testutil.JPEG(10, 10))
	require.NoError(t, err)
	_, err = command.Upload(context.Background(), "user@test.com", "two.jpg", testutil.JPEG(10, 10))
	require.NoError(t, err)

	list, err := query.ListByUser(context.Background(), "user@test.com")
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestAvatarQueryGetImageNotFound(t *testing.T) {
	query, _, _, _ := testutil.NewAvatarUseCases()
	_, err := query.GetImage(context.Background(), "missing", "", "")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestAvatarQueryGetImageStorageMissing(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	query := read.NewReadUsecase(repo, testutil.NewMemoryStorage())
	require.NoError(t, repo.Create(context.Background(), &model.Avatar{
		ID:               "avatar-1",
		UserID:           "user@test.com",
		MimeType:         imageformat.MIMEJPEG,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusCompleted,
	}))

	_, err := query.GetImage(context.Background(), "avatar-1", model.ImageSizeOriginal, "")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestAvatarQueryListByUserEmpty(t *testing.T) {
	query, _, _, _ := testutil.NewAvatarUseCases()
	list, err := query.ListByUser(context.Background(), "missing@test.com")
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestAvatarQueryGetByID(t *testing.T) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(50, 50))
	require.NoError(t, err)

	avatar, err := query.GetByID(context.Background(), resp.ID)
	require.NoError(t, err)
	assert.Equal(t, resp.ID, avatar.ID)
	assert.Equal(t, "photo.jpg", avatar.FileName)
}

func TestAvatarQueryGetUserAvatarPlaceholder(t *testing.T) {
	query, _, _, _ := testutil.NewAvatarUseCases()
	image, err := query.GetUserAvatarImage(context.Background(), "missing@test.com", model.ImageSizeSmall, imageformat.PNG)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEPNG, image.MimeType)
	assert.NotEmpty(t, image.Data)
}

func TestAvatarQueryGetUserAvatarFallbackToOriginal(t *testing.T) {
	repo := testutil.NewMemoryAvatarStore()
	store := testutil.NewMemoryStorage()
	query := read.NewReadUsecase(repo, store)

	data := testutil.JPEG(80, 80)
	avatar := &model.Avatar{
		ID:               "avatar-1",
		UserID:           "user@test.com",
		FileName:         "photo.jpg",
		MimeType:         imageformat.MIMEJPEG,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(context.Background(), avatar))
	require.NoError(t, store.SaveOriginal(context.Background(), avatar.ID, data, avatar.MimeType))

	image, err := query.GetUserAvatarImage(
		context.Background(),
		avatar.UserID,
		model.ImageSizeSmall,
		"",
	)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEJPEG, image.MimeType)
	assert.Equal(t, data, image.Data)
}

func TestAvatarQueryGetImageWithFormat(t *testing.T) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", testutil.JPEG(80, 80))
	require.NoError(t, err)

	image, err := query.GetImage(context.Background(), resp.ID, "", imageformat.PNG)
	require.NoError(t, err)
	assert.Equal(t, imageformat.MIMEPNG, image.MimeType)
	assert.NotEmpty(t, image.Data)
}

func TestAvatarQueryUploadAndGetByID(t *testing.T) {
	query, command, _, _ := testutil.NewAvatarUseCases()
	data := testutil.JPEG(640, 480)

	resp, err := command.Upload(context.Background(), "user@test.com", "photo.jpg", data)
	require.NoError(t, err)

	avatar, err := query.GetByID(context.Background(), resp.ID)
	require.NoError(t, err)
	assert.Equal(t, "photo.jpg", avatar.FileName)
	assert.Equal(t, int64(len(data)), avatar.SizeBytes)
}
