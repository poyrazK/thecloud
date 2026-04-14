package platform

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// RetryOpts configures retry behavior.
type RetryOpts struct {
	MaxAttempts int           // Total attempts (including the first). Default 3.
	BaseDelay   time.Duration // Initial delay before first retry. Default 500ms.
	MaxDelay    time.Duration // Cap on exponential growth. Default 30s.
	Multiplier  float64       // Exponent base. Default 2.0.
	// ShouldRetry is an optional predicate that returns false for errors
	// that should NOT be retried (e.g., validation errors, 4xx HTTP).
	// If nil, all non-nil errors are retried.
	ShouldRetry func(error) bool
}

func (o RetryOpts) withDefaults() RetryOpts {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 3
	}
	if o.BaseDelay <= 0 {
		o.BaseDelay = 500 * time.Millisecond
	}
	if o.MaxDelay <= 0 {
		o.MaxDelay = 30 * time.Second
	}
	if o.Multiplier <= 0 {
		o.Multiplier = 2.0
	}
	return o
}

// Retry executes fn up to opts.MaxAttempts times with exponential backoff
// and full jitter. It stops early if the context is cancelled or
// opts.ShouldRetry returns false.
func Retry(ctx context.Context, opts RetryOpts, fn func(ctx context.Context) error) error {
	opts = opts.withDefaults()

	var lastErr error
	for attempt := 0; attempt < opts.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return lastErr
			}
			return err
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// Check if this error is retryable.
		if opts.ShouldRetry != nil && !opts.ShouldRetry(lastErr) {
			return lastErr
		}

		// Don't sleep after the last attempt.
		if attempt == opts.MaxAttempts-1 {
			break
		}

		delay := backoffDelay(attempt, opts.BaseDelay, opts.MaxDelay, opts.Multiplier)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return lastErr
		case <-timer.C:
		}
	}

	return lastErr
}

// backoffDelay computes exponential backoff with full jitter:
// delay = random(0, min(maxDelay, baseDelay * multiplier^attempt))
func backoffDelay(attempt int, base, max time.Duration, mult float64) time.Duration {
	exp := math.Pow(mult, float64(attempt))
	calculated := time.Duration(float64(base) * exp)
	if calculated > max || calculated <= 0 {
		calculated = max
	}
	// Full jitter: uniform random in [0, calculated].
	return time.Duration(rand.Int64N(int64(calculated) + 1))
}
