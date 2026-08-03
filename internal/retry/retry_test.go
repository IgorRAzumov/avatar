package retry_test

import (
	"context"
	"errors"
	"testing"

	"avatar/internal/retry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBackoffSuccessFirstTry(t *testing.T) {
	calls := 0
	err := retry.WithBackoff(context.Background(), 3, func() error {
		calls++
		return nil
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithBackoffRetriesThenSucceeds(t *testing.T) {
	calls := 0
	attemptErrors := 0
	err := retry.WithBackoff(context.Background(), 3, func() error {
		calls++
		if calls < 2 {
			return errors.New("transient")
		}
		return nil
	}, func(int, error) {
		attemptErrors++
	})

	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	assert.Equal(t, 1, attemptErrors)
}

func TestWithBackoffContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sentinel := errors.New("boom")
	calls := 0
	err := retry.WithBackoff(ctx, 5, func() error {
		calls++
		return sentinel
	}, nil)

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, calls)
}

func TestWithBackoffExhausted(t *testing.T) {
	sentinel := errors.New("boom")
	err := retry.WithBackoff(context.Background(), 1, func() error {
		return sentinel
	}, nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.ErrorContains(t, err, "failed after 1 attempts")
}
