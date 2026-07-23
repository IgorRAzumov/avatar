package postgres_test

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"avatar/internal/adapter/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrationsRequiresPool(t *testing.T) {
	err := postgres.RunMigrations(context.Background(), nil, postgres.Migrations())
	require.ErrorContains(t, err, "database pool is required")
}

func TestEmbeddedMigrationsPresent(t *testing.T) {
	entries, err := fs.ReadDir(postgres.Migrations(), ".")
	require.NoError(t, err)

	var upMigrations int
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations++
		}
	}
	assert.Positive(t, upMigrations, "expected at least one embedded .up.sql migration")
}
