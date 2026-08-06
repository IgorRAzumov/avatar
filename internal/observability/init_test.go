package observability_test

import (
	"context"
	"testing"

	"avatar/internal/observability"

	"github.com/stretchr/testify/require"
)

func TestInitAllObservabilityEnabled(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:      "test",
		ServiceVersion:   "1.0.0",
		Environment:      "test",
		TracingEnabled:   true,
		LogsEnabled:      true,
		MetricsEnabled:   true,
		OTLPEndpoint:     "localhost:4317",
		TraceSampleRatio: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, runtime.Kit)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })
}
