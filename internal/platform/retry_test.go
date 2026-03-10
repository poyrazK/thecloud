package platform

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrySucceedsImmediately(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), RetryOpts{MaxAttempts: 3}, func(ctx context.Context) error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryRetriesOnFailure(t *testing.T) {
	var calls atomic.Int32
	err := Retry(context.Background(), RetryOpts{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    50 * time.Millisecond,
	}, func(ctx context.Context) error {
		n := calls.Add(1)
		if n < 3 {
			return errors.New("transient")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load())
}

func TestRetryExhaustsAttempts(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), RetryOpts{
		MaxAttempts: 2,
		BaseDelay:   10 * time.Millisecond,
	}, func(ctx context.Context) error {
		calls++
		return errors.New("permanent")
	})
	require.Error(t, err)
	assert.Equal(t, "permanent", err.Error())
	assert.Equal(t, 2, calls)
}

func TestRetryRespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Retry(ctx, RetryOpts{
		MaxAttempts: 10,
		BaseDelay:   50 * time.Millisecond,
	}, func(ctx context.Context) error {
		calls++
		if calls == 2 {
			cancel()
		}
		return errors.New("fail")
	})
	require.Error(t, err)
	assert.LessOrEqual(t, calls, 3) // might get 2 or 3 depending on timing
}

func TestRetryShouldRetryPredicate(t *testing.T) {
	permanent := errors.New("permanent error")
	calls := 0
	err := Retry(context.Background(), RetryOpts{
		MaxAttempts: 5,
		BaseDelay:   10 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			return !errors.Is(err, permanent)
		},
	}, func(ctx context.Context) error {
		calls++
		return permanent
	})
	require.ErrorIs(t, err, permanent)
	assert.Equal(t, 1, calls, "should not retry non-retryable errors")
}

func TestRetryDefaultOpts(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), RetryOpts{}, func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("fail")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls) // default MaxAttempts is 3
}

func TestBackoffDelay(t *testing.T) {
	base := 100 * time.Millisecond
	max := 5 * time.Second

	// Attempt 0: jitter in [base/2, base]
	for i := 0; i < 100; i++ {
		d := backoffDelay(0, base, max, 2.0)
		assert.GreaterOrEqual(t, d, base/2)
		assert.LessOrEqual(t, d, base)
	}

	// Attempt 3: calculated = 100ms * 2^3 = 800ms
	for i := 0; i < 100; i++ {
		d := backoffDelay(3, base, max, 2.0)
		assert.GreaterOrEqual(t, d, base/2)
		assert.LessOrEqual(t, d, 800*time.Millisecond)
	}
}
