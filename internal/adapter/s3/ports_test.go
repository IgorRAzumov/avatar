package s3_test

import (
	"avatar/internal/adapter/s3"
	"avatar/internal/domain/repository"
)

var _ repository.FilesRepository = (*s3.Storage)(nil)
