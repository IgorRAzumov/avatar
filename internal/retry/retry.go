// Package retry provides a small exponential-backoff retry helper shared by the
// in-process and RabbitMQ-based image processing paths.
package retry

import (
	"context"
	"fmt"
	"time"
)

// DefaultMaxAttempts is the default number of attempts before giving up.
const DefaultMaxAttempts = 5

// WithBackoff runs fn up to maxAttempts times, waiting 1s, 2s, 4s, ... between
// attempts. onAttemptError, if provided, is called after each failed attempt
// with the 1-based attempt number and the error. The wait is interrupted (and
// the context error returned) if ctx is cancelled.
func WithBackoff(
	ctx context.Context,
	maxAttempts int,
	fn func() error,
	onAttemptError func(attempt int, err error),
) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if onAttemptError != nil {
			onAttemptError(attempt+1, lastErr)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(1<<attempt) * time.Second):
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
