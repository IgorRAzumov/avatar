package testutil

import (
	"avatar/internal/config"
	"avatar/internal/domain/usecase/read"
	"avatar/internal/domain/usecase/write"
)

func NewAvatarUseCases() (*read.Usecase, *write.Usecase, *MemoryAvatarStore, *MemoryStorage) {
	repo := NewMemoryAvatarStore()
	store := NewMemoryStorage()
	query := read.NewReadUsecase(repo, store)
	command := write.NewWriteUsecase(repo, repo, store, NoopPublisher{}, config.DefaultMaxUploadBytes())
	return query, command, repo, store
}
