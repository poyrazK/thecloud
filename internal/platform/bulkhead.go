package platform

import (
	"context"
	"errors"
	"time"
)

// ErrBulkheadFull is returned when the bulkhead's concurrency limit is reached
// and the caller's timeout/context expires before a slot opens.
var ErrBulkheadFull = errors.New("bulkhead: concurrency limit reached")

// Bulkhead limits concurrent access to a resource using a semaphore pattern.
// It prevents one slow/failing component from consuming all available goroutines
// and cascading failure to other parts of the system.
type Bulkhead struct {
	name    string
	sem     chan struct{}
	timeout time.Duration
}

// BulkheadOpts configures a bulkhead.
type BulkheadOpts struct {
	Name        string        // Identifier for logging/metrics.
	MaxConc     int           // Maximum concurrent requests. Default 10.
	WaitTimeout time.Duration // How long to wait for a slot. Default 5s. 0 means use context deadline.
}

// NewBulkhead creates a new concurrency-limiting bulkhead.
func NewBulkhead(opts BulkheadOpts) *Bulkhead {
	if opts.MaxConc <= 0 {
		opts.MaxConc = 10
	}
	return &Bulkhead{
		name:    opts.Name,
		sem:     make(chan struct{}, opts.MaxConc),
		timeout: opts.WaitTimeout,
	}
}

// Execute runs fn within the bulkhead's concurrency limit.
// If the bulkhead is full and the wait timeout (or context) expires,
// ErrBulkheadFull is returned without calling fn.
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	if err := b.acquire(ctx); err != nil {
		return err
	}
	defer b.release()
	return fn()
}

func (b *Bulkhead) acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrBulkheadFull
	default:
	}

	if b.timeout > 0 {
		timer := time.NewTimer(b.timeout)
		defer timer.Stop()
		select {
		case b.sem <- struct{}{}:
			return nil
		case <-timer.C:
			return ErrBulkheadFull
		case <-ctx.Done():
			return ErrBulkheadFull
		}
	}
	// No explicit timeout — rely on context.
	select {
	case b.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrBulkheadFull
	}
}

func (b *Bulkhead) release() {
	<-b.sem
}

// Available returns the number of currently available slots.
func (b *Bulkhead) Available() int {
	return cap(b.sem) - len(b.sem)
}

// Name returns the bulkhead's configured name.
func (b *Bulkhead) Name() string {
	return b.name
}
