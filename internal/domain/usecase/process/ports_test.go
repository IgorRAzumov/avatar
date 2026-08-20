package process_test

import (
	domainrepo "avatar/internal/domain/repository"
	"avatar/internal/domain/usecase/process"
)

var _ domainrepo.EventPublisherRepository = (*process.AsyncPublisher)(nil)
