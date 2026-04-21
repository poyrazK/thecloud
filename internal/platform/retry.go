package platform

import (
	"context"
	"crypto/rand"
	"math"
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
	// Full jitter: uniform random in [0, calculated] using crypto/rand.
	n, err := randomInt64(int64(calculated) + 1)
	if err != nil {
		return calculated // fall back to max on error
	}
	return time.Duration(n)
}

// randomInt64 returns a uniform random int64 in [0, max) using crypto/rand.
func randomInt64(max int64) (int64, error) {
	if max <= 0 {
		return 0, nil
	}
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	// Build uint64 from bytes, then take modulo in uint64 space.
	// The result of v % max is always < max, and max <= math.MaxInt64,
	// so the final int64 conversion is always safe (G115-safe by value range).
	v := uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	mod := v % uint64(max)
	// modulus is in [0, max) where max <= math.MaxInt64, so safe.
	bound := int64(mod)
	return bound, nil
}
