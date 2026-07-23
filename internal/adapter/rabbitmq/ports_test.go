package rabbitmq_test

import (
	"avatar/internal/adapter/rabbitmq"
	domainrepo "avatar/internal/domain/repository"
)

var _ domainrepo.EventPublisherRepository = (*rabbitmq.Publisher)(nil)
