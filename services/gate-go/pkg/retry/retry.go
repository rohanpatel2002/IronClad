package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// DoWithExponentialBackoff executes the given operation with exponential backoff retries.
func DoWithExponentialBackoff(ctx context.Context, maxRetries int, initialBackoff time.Duration, maxBackoff time.Duration, operation func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled before retry: %w (last error: %v)", err, lastErr)
		}

		res, err := operation()
		if err == nil {
			return res, nil
		}

		lastErr = err

		if attempt == maxRetries {
			break
		}

		// Calculate next backoff with jitter
		jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
		sleepDuration := backoff + jitter

		if sleepDuration > maxBackoff {
			sleepDuration = maxBackoff
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled during backoff: %w (last error: %v)", ctx.Err(), lastErr)
		case <-time.After(sleepDuration):
		}

		backoff *= 2
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
