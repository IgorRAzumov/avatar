package app

import (
	"fmt"

	"avatar/internal/adapter/s3"
	"avatar/internal/config"
	domainrepo "avatar/internal/domain/repository"
)

func NewFileStore(cfg *config.Config) (domainrepo.FilesRepository, error) {
	store, err := s3.NewStorage(cfg.S3)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: %w", err)
	}
	return store, nil
}
