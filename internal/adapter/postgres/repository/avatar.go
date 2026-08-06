package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"avatar/internal/adapter/objectkey"
	"avatar/internal/domain/model"
	"avatar/internal/observability"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAvatarStore struct {
	pool *pgxpool.Pool
	kit  observability.Kit
}

func NewPostgresAvatarStore(pool *pgxpool.Pool, kit observability.Kit) *PostgresAvatarStore {
	return &PostgresAvatarStore{pool: pool, kit: kit}
}

func (store *PostgresAvatarStore) Close() {
	if store.pool != nil {
		store.pool.Close()
	}
}

func (store *PostgresAvatarStore) Ping(ctx context.Context) error {
	return store.withSpan(ctx, "db.ping", func(ctx context.Context) error {
		return store.pool.Ping(ctx)
	})
}

func (store *PostgresAvatarStore) Create(ctx context.Context, avatar *model.Avatar) error {
	return store.withSpan(ctx, "db.create_avatar", func(ctx context.Context) error {
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
	}, observability.String("user_id", avatar.UserID))
}

func (store *PostgresAvatarStore) GetByID(ctx context.Context, id string) (*model.Avatar, error) {
	var avatar *model.Avatar
	err := store.withSpan(ctx, "db.get_avatar_by_id", func(ctx context.Context) error {
		query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE id = $1 AND deleted_at IS NULL`

		result, queryErr := store.scanAvatar(ctx, query, id)
		if queryErr != nil {
			return queryErr
		}
		avatar = result
		return nil
	}, observability.String("avatar_id", id))
	return avatar, err
}

func (store *PostgresAvatarStore) GetLatestByUserID(ctx context.Context, userID string) (*model.Avatar, error) {
	var avatar *model.Avatar
	err := store.withSpan(ctx, "db.get_latest_avatar", func(ctx context.Context) error {
		query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

		result, queryErr := store.scanAvatar(ctx, query, userID)
		if queryErr != nil {
			return queryErr
		}
		avatar = result
		return nil
	}, observability.String("user_id", userID))
	return avatar, err
}

func (store *PostgresAvatarStore) ListByUserID(ctx context.Context, userID string) ([]*model.Avatar, error) {
	var avatars []*model.Avatar
	err := store.withSpan(ctx, "db.list_avatars", func(ctx context.Context) error {
		query := `
		SELECT id, user_id, file_name, mime_type, size_bytes,
		       upload_status, processing_status, width, height, created_at, updated_at, deleted_at
		FROM avatars
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

		rows, queryErr := store.pool.Query(ctx, query, userID)
		if queryErr != nil {
			return fmt.Errorf("query avatars: %w", queryErr)
		}
		defer rows.Close()

		for rows.Next() {
			a, scanErr := scanAvatarRow(rows)
			if scanErr != nil {
				return scanErr
			}
			avatars = append(avatars, a)
		}
		return rows.Err()
	}, observability.String("user_id", userID))
	return avatars, err
}

func (store *PostgresAvatarStore) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := store.withSpan(ctx, "db.count_active_avatars", func(ctx context.Context) error {
		return store.pool.QueryRow(ctx, `SELECT COUNT(*) FROM avatars WHERE deleted_at IS NULL`).Scan(&count)
	})
	return count, err
}

func (store *PostgresAvatarStore) CountActiveByUser(ctx context.Context) (map[string]int64, error) {
	counts := make(map[string]int64)
	err := store.withSpan(ctx, "db.count_active_avatars_by_user", func(ctx context.Context) error {
		rows, queryErr := store.pool.Query(ctx, `
			SELECT user_id, COUNT(*)
			FROM avatars
			WHERE deleted_at IS NULL
			GROUP BY user_id`)
		if queryErr != nil {
			return fmt.Errorf("query avatar counts: %w", queryErr)
		}
		defer rows.Close()

		for rows.Next() {
			var userID string
			var count int64
			if scanErr := rows.Scan(&userID, &count); scanErr != nil {
				return scanErr
			}
			counts[userID] = count
		}
		return rows.Err()
	})
	return counts, err
}

func (store *PostgresAvatarStore) SoftDelete(ctx context.Context, id string) error {
	return store.withSpan(ctx, "db.soft_delete_avatar", func(ctx context.Context) error {
		query := `UPDATE avatars SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
		tag, err := store.pool.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("soft delete: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return model.ErrNotFound
		}
		return nil
	}, observability.String("avatar_id", id))
}

func (store *PostgresAvatarStore) UpdateProcessingStatus(ctx context.Context, id, status string) error {
	return store.withSpan(ctx, "db.update_processing_status", func(ctx context.Context) error {
		query := `UPDATE avatars SET processing_status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
		_, err := store.pool.Exec(ctx, query, id, status)
		return err
	}, observability.String("avatar_id", id), observability.String("status", status))
}

func (store *PostgresAvatarStore) CompleteProcessing(ctx context.Context, id string, width, height int) error {
	return store.withSpan(ctx, "db.complete_processing", func(ctx context.Context) error {
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
	}, observability.String("avatar_id", id))
}

func (store *PostgresAvatarStore) withSpan(
	ctx context.Context,
	operation string,
	fn func(context.Context) error,
	attrs ...observability.Attr,
) error {
	return store.kit.RunDB(ctx, operation, fn, attrs...)
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
