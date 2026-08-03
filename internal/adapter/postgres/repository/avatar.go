package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"avatar/internal/adapter/objectkey"
	"avatar/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAvatarStore struct {
	pool *pgxpool.Pool
}

func NewPostgresAvatarStore(pool *pgxpool.Pool) *PostgresAvatarStore {
	return &PostgresAvatarStore{pool: pool}
}

func (store *PostgresAvatarStore) Close() {
	if store.pool != nil {
		store.pool.Close()
	}
}

func (store *PostgresAvatarStore) Ping(ctx context.Context) error {
	return store.pool.Ping(ctx)
}

func (store *PostgresAvatarStore) Create(ctx context.Context, avatar *model.Avatar) error {
	if avatar.ID == "" {
		avatar.ID = uuid.New().String()
	}

	query := `
		INSERT INTO avatars (id, user_id, file_name, mime_type, size_bytes, s3_key, thumbnail_s3_keys, upload_status, processing_status, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	return store.pool.QueryRow(ctx, query,
		avatar.ID, avatar.UserID, avatar.FileName, avatar.MimeType,
		avatar.SizeBytes, objectkey.OriginalObjectKey(avatar.ID), []byte("{}"),
		avatar.UploadStatus, avatar.ProcessingStatus,
		avatar.Width, avatar.Height,
	).Scan(&avatar.CreatedAt, &avatar.UpdatedAt)
}

func (store *PostgresAvatarStore) GetByID(ctx context.Context, id string) (*model.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE id = $1 AND deleted_at IS NULL`

	return store.scanAvatar(ctx, query, id)
}

func (store *PostgresAvatarStore) GetLatestByUserID(ctx context.Context, userID string) (*model.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	return store.scanAvatar(ctx, query, userID)
}

func (store *PostgresAvatarStore) ListByUserID(ctx context.Context, userID string) ([]*model.Avatar, error) {
	query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := store.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query avatars: %w", err)
	}
	defer rows.Close()

	var avatars []*model.Avatar
	for rows.Next() {
		a, err := scanAvatarRow(rows)
		if err != nil {
			return nil, err
		}
		avatars = append(avatars, a)
	}
	return avatars, rows.Err()
}

func (store *PostgresAvatarStore) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE avatars SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := store.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (store *PostgresAvatarStore) UpdateProcessingStatus(ctx context.Context, id, status string) error {
	query := `UPDATE avatars SET processing_status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := store.pool.Exec(ctx, query, id, status)
	return err
}

func (store *PostgresAvatarStore) CompleteProcessing(ctx context.Context, id string, width, height int) error {
	thumbsJSON, err := json.Marshal(thumbnailKeysForDB(id))
	if err != nil {
		return fmt.Errorf("marshal thumbnails: %w", err)
	}

	query := `
		UPDATE avatars
		SET thumbnail_s3_keys = $2, processing_status = $3, width = $4, height = $5, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	_, err = store.pool.Exec(ctx, query, id, thumbsJSON, model.ProcessingStatusCompleted, width, height)
	return err
}

func thumbnailKeysForDB(avatarID string) map[string]string {
	keys := make(map[string]string, len(model.ThumbnailSizes))
	for _, size := range model.ThumbnailSizes {
		keys[size.Name] = objectkey.ThumbnailObjectKey(avatarID, size.Name)
	}
	return keys
}

func (store *PostgresAvatarStore) scanAvatar(ctx context.Context, query string, args ...any) (*model.Avatar, error) {
	row := store.pool.QueryRow(ctx, query, args...)
	a, err := scanAvatarFromRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAvatarFromRow(row scannable) (*model.Avatar, error) {
	var a model.Avatar
	var deletedAt *time.Time

	err := row.Scan(
		&a.ID, &a.UserID, &a.FileName, &a.MimeType, &a.SizeBytes,
		&a.UploadStatus, &a.ProcessingStatus,
		&a.Width, &a.Height, &a.CreatedAt, &a.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	a.DeletedAt = deletedAt
	return &a, nil
}

func scanAvatarRow(rows pgx.Rows) (*model.Avatar, error) {
	return scanAvatarFromRow(rows)
}
