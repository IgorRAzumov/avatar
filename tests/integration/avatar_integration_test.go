//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"avatar/internal/domain/usecase/read"
	"avatar/internal/domain/usecase/write"

	avatarpg "avatar/internal/adapter/postgres"
	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/config"
	"avatar/internal/domain/model"
	"avatar/internal/imageformat"
	"avatar/internal/logger"
	"avatar/internal/processor"
	"avatar/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("avatar"),
		postgres.WithUsername("avatar"),
		postgres.WithPassword("avatar"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}
	return pool, cleanup
}

func TestRepositoryCreateAndGet(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, avatarpg.RunMigrations(ctx, pool, avatarpg.Migrations()))

	repo := repository.NewPostgresAvatarStore(pool)
	avatar := &model.Avatar{
		UserID:           "user@test.com",
		FileName:         "photo.jpg",
		MimeType:         imageformat.MIMEJPEG,
		SizeBytes:        128,
		UploadStatus:     model.UploadStatusCompleted,
		ProcessingStatus: model.ProcessingStatusPending,
	}
	require.NoError(t, repo.Create(ctx, avatar))

	got, err := repo.GetByID(ctx, avatar.ID)
	require.NoError(t, err)
	require.Equal(t, avatar.UserID, got.UserID)

	require.NoError(t, repo.SoftDelete(ctx, avatar.ID))
	_, err = repo.GetByID(ctx, avatar.ID)
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestServiceUploadFlow(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, avatarpg.RunMigrations(ctx, pool, avatarpg.Migrations()))

	store := testutil.NewMemoryStorage()

	repo := repository.NewPostgresAvatarStore(pool)
	proc := processor.NewAsyncImageResizerProcessor(repo, repo, store, logger.Nop())
	query := read.NewReadUsecase(repo, store)
	command := write.NewWriteUsecase(repo, repo, store, proc, config.DefaultMaxUploadBytes())

	resp, err := command.Upload(ctx, "user@test.com", "photo.jpg", testutil.JPEG(120, 120))
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	time.Sleep(200 * time.Millisecond)

	image, err := query.GetImage(ctx, resp.ID, model.ImageSizeSmall, "")
	require.NoError(t, err)
	require.Equal(t, imageformat.MIMEJPEG, image.MimeType)
	require.NotEmpty(t, image.Data)
}
