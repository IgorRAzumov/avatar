package health

import (
	"context"

	"avatar/internal/domain/model"
	domainrepo "avatar/internal/domain/repository"
)

const defaultStorageComponent = "storage"

type pinger interface {
	Ping(ctx context.Context) error
}

// Checker aggregates liveness of the service dependencies (database, file
// storage and, optionally, the message broker) into a single health report.
type Checker struct {
	database         pinger
	storage          pinger
	broker           pinger
	storageComponent string
}

func NewChecker(
	database pinger,
	files domainrepo.FilesRepository,
	storageComponent string,
	broker pinger,
) *Checker {
	if storageComponent == "" {
		storageComponent = defaultStorageComponent
	}

	var storage pinger
	if p, ok := files.(pinger); ok {
		storage = p
	}
	return &Checker{
		database:         database,
		storage:          storage,
		broker:           broker,
		storageComponent: storageComponent,
	}
}

func (checker *Checker) Check(ctx context.Context) model.HealthReport {
	components := map[string]string{
		"database":               model.ComponentStatusOK,
		checker.storageComponent: model.ComponentStatusOK,
	}
	status := model.HealthStatusOK

	if !ping(ctx, checker.database) {
		components["database"] = model.ComponentStatusError
		status = model.HealthStatusDegraded
	}
	if !ping(ctx, checker.storage) {
		components[checker.storageComponent] = model.ComponentStatusError
		status = model.HealthStatusDegraded
	}
	if checker.broker != nil {
		components["broker"] = model.ComponentStatusOK
		if err := checker.broker.Ping(ctx); err != nil {
			components["broker"] = model.ComponentStatusError
			status = model.HealthStatusDegraded
		}
	}

	return model.HealthReport{Status: status, Components: components}
}

func ping(ctx context.Context, p pinger) bool {
	if p == nil {
		return false
	}
	return p.Ping(ctx) == nil
}
