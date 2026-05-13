package postgres

import (
	"context"
	"fmt"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	Backoff     []time.Duration
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Backoff:     []time.Duration{50 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond},
}

func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			backoffIdx := attempt - 1
			if backoffIdx >= len(cfg.Backoff) {
				backoffIdx = len(cfg.Backoff) - 1
			}
			select {
			case <-time.After(cfg.Backoff[backoffIdx]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		if !isRetryableDBError(err) {
			return err
		}

		lastErr = err
	}
	return fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

func isRetryableDBError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	retryable := []string{
		"connection refused",
		"connection reset",
		"too many connections",
		"deadlock detected",
		"serialization failure",
		"could not serialize",
		"server closed the connection",
		"i/o timeout",
	}
	for _, pattern := range retryable {
		if contains(msg, pattern) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
