package health_test

import (
	"context"
	"errors"
	"testing"

	domainrepo "avatar/internal/domain/repository"
	"avatar/internal/domain/usecase/health"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
)

type failPinger struct{}

func (failPinger) Ping(context.Context) error { return errors.New("down") }

func TestHealthCheckerOK(t *testing.T) {
	checker := health.NewChecker(testutil.NewMemoryAvatarStore(), testutil.NewMemoryStorage(), "storage", nil)
	resp := checker.Check(context.Background())
	assert.Equal(t, "ok", resp.Status)
}

func TestHealthCheckerS3Component(t *testing.T) {
	checker := health.NewChecker(testutil.NewMemoryAvatarStore(), testutil.NewMemoryStorage(), "s3", nil)
	resp := checker.Check(context.Background())
	assert.Equal(t, "ok", resp.Components["s3"])
}

func TestHealthCheckerBrokerDown(t *testing.T) {
	checker := health.NewChecker(testutil.NewMemoryAvatarStore(), testutil.NewMemoryStorage(), "storage", failPinger{})
	resp := checker.Check(context.Background())
	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "error", resp.Components["broker"])
}

func TestHealthCheckerDatabaseDown(t *testing.T) {
	checker := health.NewChecker(failPinger{}, testutil.NewMemoryStorage(), "storage", nil)
	resp := checker.Check(context.Background())
	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "error", resp.Components["database"])
}

var _ domainrepo.HealthRepository = (*health.Checker)(nil)
