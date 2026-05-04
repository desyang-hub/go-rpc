// Package retry provides exponential backoff retry mechanism.
//
// The retry package supports two modes:
//  1. Retry  - Retry up to N times with backoff
//  2. RetryUntilFlood - Retry until success or budget exhausted
//
// Usage:
//
//	retry := retry.NewWithPolicy(retry.DefaultPolicy)
//	result, err := retry.Do(func(ctx context.Context) (string, error) {
//	    return httpClient.Get("https://example.com")
//	}, ctx)
package retry

import (
	"context"
	"time"
)

// Config holds retry configuration.
type Config struct {
	// Maximum number of retry attempts (0 = no retry)
	MaxAttempts int

	// Initial backoff duration
	InitialBackoff time.Duration

	// Maximum backoff duration cap
	MaxBackoff time.Duration

	// If true, backoff doubles each attempt
	ExponentialBackoff bool

	// Additional delay after backoff calculation
	FixedDelay time.Duration

	// Maximum duration to allow for retries (stops retry if exceeded)
	MaxRetryingTime time.Duration
}

// DefaultRetryConfig returns a retry configuration with 3 attempts and exponential backoff.
func DefaultRetryConfig() Config {
	return Config{
		MaxAttempts:        3,
		InitialBackoff:     100 * time.Millisecond,
		MaxBackoff:         3 * time.Second,
		ExponentialBackoff: true,
	}
}

// DoFn is the function to retry.
type DoFn func(ctx context.Context) error

// Do executes the function with retry logic.
//
// If the function succeeds, the error returned is nil.
// If the function fails and no more retries are available, the last error
// is returned.
//
// If ctx is cancelled before all retries are exhausted, ctx.Done() error is returned.
func Do(ctx context.Context, fn DoFn, cfg Config) error {
	var lastErr error

	// If MaxRetryingTime is set, create a budget timer
	var budgetTimer *time.Timer
	if cfg.MaxRetryingTime > 0 {
		budgetTimer = time.NewTimer(cfg.MaxRetryingTime)
		defer budgetTimer.Stop()
	}

	retryBudget := func() bool {
		if budgetTimer != nil {
			if budgetTimer.Stop() {
				return true
			}
			select {
			case <-budgetTimer.C:
				return false
			default:
				return true
			}
		}
		return true
	}

	for attempt := 0; attempt <= cfg.MaxAttempts; attempt++ {
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// No more retries available
		if attempt == cfg.MaxAttempts {
			break
		}

		// Check context before waiting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check retry budget
		if !retryBudget() {
			return ctx.Err()
		}

		// Calculate backoff duration
		backoff := cfg.InitialBackoff

		if cfg.ExponentialBackoff {
			// Double the backoff for each attempt
			for i := 0; i < attempt; i++ {
				backoff *= 2
			}
		}

		// Apply fixed delay
		backoff += cfg.FixedDelay

		// Cap at maximum backoff
		if cfg.MaxBackoff > 0 && backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	return lastErr
}
