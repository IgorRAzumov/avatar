package s3_test

import (
	"context"
	"testing"

	"avatar/internal/adapter/s3"
	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageAvatarFiles(t *testing.T) {
	store := startMinIO(t)
	ctx := context.Background()

	require.NoError(t, store.SaveOriginal(ctx, "avatar-1", []byte("image"), "image/jpeg"))

	data, err := store.OpenOriginal(ctx, "avatar-1")
	require.NoError(t, err)
	assert.Equal(t, []byte("image"), data)

	require.NoError(t, store.SaveThumbnail(ctx, "avatar-1", model.ImageSizeSmall, []byte("thumb")))
	thumb, err := store.OpenThumbnail(ctx, "avatar-1", model.ImageSizeSmall)
	require.NoError(t, err)
	assert.Equal(t, []byte("thumb"), thumb)

	require.NoError(t, store.DeleteAll(ctx, "avatar-1"))
	_, err = store.OpenOriginal(ctx, "avatar-1")
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestStoragePing(t *testing.T) {
	store := startMinIO(t)
	require.NoError(t, store.Ping(context.Background()))
}

func startMinIO(t *testing.T) *s3.Storage {
	t.Helper()

	endpoint := "127.0.0.1:9000"
	store, err := s3.NewStorage(config.S3Config{
		Endpoint:  endpoint,
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Bucket:    "avatars",
		UseSSL:    false,
		Region:    "us-east-1",
	}, observability.NewTestKit())
	if err != nil {
		t.Skipf("minio not available: %v", err)
	}

	ctx := context.Background()
	if err := store.Ping(ctx); err != nil {
		t.Skipf("minio not available at %s: %v", endpoint, err)
	}

	return store
}
