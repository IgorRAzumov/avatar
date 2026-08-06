package app

import (
	"fmt"

	"avatar/internal/adapter/s3"
	"avatar/internal/config"
	domainrepo "avatar/internal/domain/repository"
	"avatar/internal/observability"
)

func NewFileStore(cfg *config.Config, kit observability.Kit) (domainrepo.FilesRepository, error) {
	store, err := s3.NewStorage(cfg.S3, kit)
	if err != nil {
		return nil, fmt.Errorf("create s3 storage: %w", err)
	}
	return store, nil
}
