package observability_test

import (
	"context"
	"errors"
	"testing"

	"avatar/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRecordsError(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		TracingEnabled: true,
		MetricsEnabled: false,
		LogsEnabled:    false,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })

	testErr := errors.New("boom")
	err = runtime.Kit.Run(context.Background(), "test.run", func(context.Context) error {
		return testErr
	})
	assert.ErrorIs(t, err, testErr)
}

func TestRunResultReturnsValue(t *testing.T) {
	runtime, err := observability.Init(context.Background(), observability.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		TracingEnabled: true,
		MetricsEnabled: false,
		LogsEnabled:    false,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = runtime.Shutdown(context.Background()) })

	value, err := observability.RunResult(runtime.Kit, context.Background(), "test.run_result", func(context.Context) (string, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", value)
}
