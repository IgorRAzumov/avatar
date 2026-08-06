package postgres

import (
	"context"
	"fmt"

	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/observability"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, dsn string, kit observability.Kit) (*repository.PostgresAvatarStore, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := RunMigrations(ctx, pool, Migrations()); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("migrate: %w", err)
	}

	store := repository.NewPostgresAvatarStore(pool, kit)

	metricsCleanup, err := kit.RegisterDatabaseMetrics(observability.DatabaseStats{
		PoolConnections: func() (int64, int64) {
			stat := pool.Stat()
			return int64(stat.AcquiredConns()), int64(stat.IdleConns())
		},
		CountAvatars: countActiveAvatars(pool),
	})
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("register db metrics: %w", err)
	}

	closeFn := func() {
		if metricsCleanup != nil {
			metricsCleanup()
		}
		pool.Close()
	}

	return store, closeFn, nil
}

func countActiveAvatars(pool *pgxpool.Pool) func(context.Context) (int64, map[string]int64, error) {
	return func(ctx context.Context) (int64, map[string]int64, error) {
		var total int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM avatars WHERE deleted_at IS NULL`).Scan(&total); err != nil {
			return 0, nil, fmt.Errorf("count avatars: %w", err)
		}

		rows, err := pool.Query(ctx, `
			SELECT user_id, COUNT(*)
			FROM avatars
			WHERE deleted_at IS NULL
			GROUP BY user_id`)
		if err != nil {
			return 0, nil, fmt.Errorf("count avatars by user: %w", err)
		}
		defer rows.Close()

		byUser := make(map[string]int64)
		for rows.Next() {
			var userID string
			var count int64
			if err := rows.Scan(&userID, &count); err != nil {
				return 0, nil, err
			}
			byUser[userID] = count
		}
		if err := rows.Err(); err != nil {
			return 0, nil, err
		}

		return total, byUser, nil
	}
}
