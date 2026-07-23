package postgres

import (
	"context"
	"fmt"

	"avatar/internal/adapter/postgres/repository"

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

	return repository.NewPostgresAvatarStore(pool), nil
}
