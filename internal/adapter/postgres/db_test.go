package postgres_test

import (
	"testing"

	"avatar/internal/adapter/postgres/repository"
)

func TestNewPostgresAvatarStore(t *testing.T) {
	if repository.NewPostgresAvatarStore(nil) == nil {
		t.Fatal("expected store")
	}
}
