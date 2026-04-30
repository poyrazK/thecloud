package platform

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkheadAllowsUpToMaxConcurrency(t *testing.T) {
	bh := NewBulkhead(BulkheadOpts{Name: "test", MaxConc: 2})

	var running atomic.Int32
	var maxSeen atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bh.Execute(context.Background(), func() error {
				cur := running.Add(1)
				defer running.Add(-1)
				// Track the max concurrent.
				for {
					old := maxSeen.Load()
					if cur <= old || maxSeen.CompareAndSwap(old, cur) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, maxSeen.Load(), int32(2))
}

func TestBulkheadRejectsWhenFull(t *testing.T) {
	bh := NewBulkhead(BulkheadOpts{Name: "test", MaxConc: 1, WaitTimeout: 50 * time.Millisecond})

	// Fill the bulkhead.
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = bh.Execute(context.Background(), func() error {
			close(started)
			<-done
			return nil
		})
	}()
	<-started

	// Second call should be rejected.
	err := bh.Execute(context.Background(), func() error { return nil })
	require.ErrorIs(t, err, ErrBulkheadFull)

	close(done)
}

func TestBulkheadRespectsContext(t *testing.T) {
	bh := NewBulkhead(BulkheadOpts{Name: "test", MaxConc: 1})

	// Fill the bulkhead.
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = bh.Execute(context.Background(), func() error {
			close(started)
			<-done
			return nil
		})
	}()
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := bh.Execute(ctx, func() error { return nil })
	require.ErrorIs(t, err, context.DeadlineExceeded)

	close(done)
}

func TestBulkheadPropagatesFunctionError(t *testing.T) {
	bh := NewBulkhead(BulkheadOpts{Name: "test", MaxConc: 5})
	myErr := errors.New("business error")
	err := bh.Execute(context.Background(), func() error { return myErr })
	require.ErrorIs(t, err, myErr)
}

func TestBulkheadAvailable(t *testing.T) {
	bh := NewBulkhead(BulkheadOpts{MaxConc: 3})
	assert.Equal(t, 3, bh.Available())

	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		_ = bh.Execute(context.Background(), func() error {
			close(started)
			<-done
			return nil
		})
	}()
	<-started
	assert.Equal(t, 2, bh.Available())
	close(done)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 3, bh.Available())
}
