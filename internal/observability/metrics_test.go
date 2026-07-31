package observability_test

import (
	"context"
	"testing"

	"avatar/internal/observability"

	"github.com/stretchr/testify/require"
)

func TestRegisterDatabaseMetricsObservesSource(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		MetricsEnabled: true,
		TracingEnabled: false,
		LogsEnabled:    false,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })

	unregister, err := observability.RegisterDatabaseMetrics(observability.DatabaseStats{
		PoolConnections: func() (int64, int64) { return 2, 3 },
		CountAvatars: func(context.Context) (int64, map[string]int64, error) {
			return 4, map[string]int64{"user-a": 2, "user-b": 2}, nil
		},
	})
	require.NoError(t, err)
	require.NotNil(t, unregister)

	unregister()
}

func TestRegisterDatabaseMetricsNoopWhenMetricsDisabled(t *testing.T) {
	unregister, err := observability.RegisterDatabaseMetrics(observability.DatabaseStats{
		PoolConnections: func() (int64, int64) { return 1, 1 },
		CountAvatars: func(context.Context) (int64, map[string]int64, error) {
			return 1, nil, nil
		},
	})
	require.NoError(t, err)
	require.NotNil(t, unregister)
	unregister()
}
