package repository_test

import (
	"avatar/internal/adapter/postgres/repository"
	domainrepo "avatar/internal/domain/repository"
)

var (
	_ domainrepo.ReadRepository  = (*repository.PostgresAvatarStore)(nil)
	_ domainrepo.WriteRepository = (*repository.PostgresAvatarStore)(nil)
)
