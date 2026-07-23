package processor_test

import (
	domainrepo "avatar/internal/domain/repository"
	"avatar/internal/processor"
)

var _ domainrepo.EventPublisherRepository = (*processor.AsyncPublisher)(nil)
