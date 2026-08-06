package postgres_test

import (
	"testing"

	"avatar/internal/adapter/postgres/repository"
	"avatar/internal/observability"
)

func TestNewPostgresAvatarStore(t *testing.T) {
	if repository.NewPostgresAvatarStore(nil, observability.NewTestKit()) == nil {
		t.Fatal("expected store")
	}
}
