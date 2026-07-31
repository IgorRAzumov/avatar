package postgres

import (
	"context"
	"fmt"

	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/observability"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, dsn string) (*repository.PostgresAvatarStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := RunMigrations(ctx, pool, Migrations()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	store := repository.NewPostgresAvatarStore(pool)

	cleanup, err := observability.RegisterDatabaseMetrics(observability.DatabaseStats{
		PoolConnections: func() (int64, int64) {
			stat := pool.Stat()
			return int64(stat.AcquiredConns()), int64(stat.IdleConns())
		},
		CountAvatars: func(ctx context.Context) (int64, map[string]int64, error) {
			total, countErr := store.CountActive(ctx)
			if countErr != nil {
				return 0, nil, countErr
			}

			byUser, countErr := store.CountActiveByUser(ctx)
			if countErr != nil {
				return 0, nil, countErr
			}

			return total, byUser, nil
		},
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("register db metrics: %w", err)
	}
	store.SetMetricsCleanup(cleanup)

	return store, nil
}
