// Package drills provides integration-like failure drill tests that validate
// the HA properties of the control plane. These tests use mocks to simulate
// infrastructure failures without requiring real Postgres/Redis.
//
// Run: go test ./internal/drills/ -v -count=1
package drills

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/platform"
)

// ---------------------------------------------------------------------------
// Drill 1: Circuit breaker trip + recovery
// SLO: When a backend fails ≥ threshold times, all subsequent calls must
// return ErrCircuitOpen within 1ms (no backend call). After resetTimeout,
// a successful probe must close the circuit.
// ---------------------------------------------------------------------------

func TestDrill_CircuitBreakerTripAndRecovery(t *testing.T) {
	const threshold = 3
	const resetTimeout = 200 * time.Millisecond

	var transitions []string
	var mu sync.Mutex

	cb := platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
		Name:            "drill-cb",
		Threshold:       threshold,
		ResetTimeout:    resetTimeout,
		SuccessRequired: 1,
		OnStateChange: func(name string, from, to platform.State) {
			mu.Lock()
			transitions = append(transitions, from.String()+"→"+to.String())
			mu.Unlock()
		},
	})

	backendErr := errors.New("backend down")

	// Phase 1: Trip the circuit with consecutive failures.
	for i := 0; i < threshold; i++ {
		err := cb.Execute(func() error { return backendErr })
		if err == nil {
			t.Fatalf("iteration %d: expected error", i)
		}
	}

	// Verify circuit is open.
	if cb.GetState() != platform.StateOpen {
		t.Fatalf("expected open, got %s", cb.GetState().String())
	}

	// Phase 2: Confirm fail-fast (no backend call).
	var backendCalled atomic.Bool
	start := time.Now()
	err := cb.Execute(func() error {
		backendCalled.Store(true)
		return nil
	})
	elapsed := time.Since(start)

	if !errors.Is(err, platform.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if backendCalled.Load() {
		t.Fatal("backend should NOT have been called while circuit is open")
	}
	if elapsed > 5*time.Millisecond {
		t.Fatalf("fail-fast took %v, expected <5ms", elapsed)
	}

	// Phase 3: Wait for resetTimeout, then recover.
	time.Sleep(resetTimeout + 50*time.Millisecond)
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
	if cb.GetState() != platform.StateClosed {
		t.Fatalf("expected closed after recovery, got %s", cb.GetState().String())
	}

	// Verify transitions: closed→open, open→half-open, half-open→closed.
	mu.Lock()
	defer mu.Unlock()
	expected := []string{"closed→open", "open→half-open", "half-open→closed"}
	if len(transitions) != len(expected) {
		t.Fatalf("expected %d transitions, got %d: %v", len(expected), len(transitions), transitions)
	}
	for i := range expected {
		if transitions[i] != expected[i] {
			t.Fatalf("transition[%d]: expected %s, got %s", i, expected[i], transitions[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Drill 2: Bulkhead saturation + graceful rejection
// SLO: When maxConc requests are in-flight, additional requests must be
// rejected with ErrBulkheadFull (not blocked forever).
// ---------------------------------------------------------------------------

func TestDrill_BulkheadSaturationAndRejection(t *testing.T) {
	const maxConc = 3
	const waitTimeout = 100 * time.Millisecond

	bh := platform.NewBulkhead(platform.BulkheadOpts{
		Name:        "drill-bh",
		MaxConc:     maxConc,
		WaitTimeout: waitTimeout,
	})

	ctx := context.Background()
	blockCh := make(chan struct{})
	var inFlight atomic.Int64
	var rejected atomic.Int64
	var wg sync.WaitGroup

	// Saturate the bulkhead.
	for i := 0; i < maxConc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bh.Execute(ctx, func() error {
				inFlight.Add(1)
				<-blockCh
				return nil
			})
		}()
	}

	// Wait for all slots to be occupied.
	for inFlight.Load() < int64(maxConc) {
		time.Sleep(5 * time.Millisecond)
	}

	if bh.Available() != 0 {
		t.Fatalf("expected 0 available slots, got %d", bh.Available())
	}

	// Fire excess requests — they should be rejected.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bh.Execute(ctx, func() error { return nil })
			if errors.Is(err, platform.ErrBulkheadFull) {
				rejected.Add(1)
			}
		}()
	}

	// Let the excess timeout.
	time.Sleep(waitTimeout + 50*time.Millisecond)

	// Unblock the saturating goroutines.
	close(blockCh)
	wg.Wait()

	if rejected.Load() != 5 {
		t.Fatalf("expected 5 rejections, got %d", rejected.Load())
	}
}

// ---------------------------------------------------------------------------
// Drill 3: Resilient adapter end-to-end (circuit + bulkhead + timeout)
// SLO: A failing backend trips the circuit; subsequent calls fail-fast;
// recovery probe succeeds and normal operation resumes.
// ---------------------------------------------------------------------------

type failingBackend struct {
	healthy atomic.Bool
	calls   atomic.Int64
}

func (f *failingBackend) Do(_ context.Context) error {
	f.calls.Add(1)
	if f.healthy.Load() {
		return nil
	}
	return errors.New("backend failure")
}

func TestDrill_ResilientAdapterEndToEnd(t *testing.T) {
	backend := &failingBackend{}
	logger := slog.Default()

	cb := platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
		Name:            "drill-e2e",
		Threshold:       3,
		ResetTimeout:    200 * time.Millisecond,
		SuccessRequired: 1,
	})

	bh := platform.NewBulkhead(platform.BulkheadOpts{
		Name:    "drill-e2e",
		MaxConc: 5,
	})

	_ = logger // Would be used for real logging in production.

	// Helper: simulate calling through the full resilience stack.
	callThrough := func(ctx context.Context) error {
		return bh.Execute(ctx, func() error {
			return cb.Execute(func() error {
				ctx2, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
				defer cancel()
				return backend.Do(ctx2)
			})
		})
	}

	ctx := context.Background()

	// Phase 1: Backend is down → trip circuit.
	for i := 0; i < 3; i++ {
		_ = callThrough(ctx)
	}
	if cb.GetState() != platform.StateOpen {
		t.Fatalf("expected open, got %s", cb.GetState().String())
	}

	// Phase 2: Fail-fast while open.
	callsBefore := backend.calls.Load()
	err := callThrough(ctx)
	if !errors.Is(err, platform.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if backend.calls.Load() != callsBefore {
		t.Fatal("backend should not be called while circuit is open")
	}

	// Phase 3: Backend recovers.
	backend.healthy.Store(true)
	time.Sleep(250 * time.Millisecond)

	err = callThrough(ctx)
	if err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
	if cb.GetState() != platform.StateClosed {
		t.Fatalf("expected closed, got %s", cb.GetState().String())
	}

	// Phase 4: Normal operation continues.
	for i := 0; i < 10; i++ {
		if err := callThrough(ctx); err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Drill 4: Retry with exponential backoff
// SLO: Retry must respect MaxAttempts, must apply backoff between attempts,
// and must stop early if the context is cancelled.
// ---------------------------------------------------------------------------

func TestDrill_RetryBackoffAndContextCancellation(t *testing.T) {
	t.Run("exhausts_attempts", func(t *testing.T) {
		var attempts atomic.Int64
		err := platform.Retry(context.Background(), platform.RetryOpts{
			MaxAttempts: 4,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    50 * time.Millisecond,
		}, func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("still failing")
		})

		if err == nil {
			t.Fatal("expected error after exhausting retries")
		}
		if attempts.Load() != 4 {
			t.Fatalf("expected 4 attempts, got %d", attempts.Load())
		}
	})

	t.Run("stops_on_context_cancel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		var attempts atomic.Int64
		start := time.Now()
		err := platform.Retry(ctx, platform.RetryOpts{
			MaxAttempts: 100, // Would take very long if not cancelled.
			BaseDelay:   50 * time.Millisecond,
			MaxDelay:    200 * time.Millisecond,
		}, func(ctx context.Context) error {
			attempts.Add(1)
			return errors.New("failing")
		})

		elapsed := time.Since(start)
		if err == nil {
			t.Fatal("expected error")
		}
		if elapsed > 500*time.Millisecond {
			t.Fatalf("should have stopped early, took %v", elapsed)
		}
		if attempts.Load() >= 100 {
			t.Fatal("should not have exhausted all 100 attempts")
		}
	})

	t.Run("succeeds_on_retry", func(t *testing.T) {
		var attempts atomic.Int64
		err := platform.Retry(context.Background(), platform.RetryOpts{
			MaxAttempts: 5,
			BaseDelay:   5 * time.Millisecond,
		}, func(ctx context.Context) error {
			n := attempts.Add(1)
			if n < 3 {
				return errors.New("not yet")
			}
			return nil
		})

		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if attempts.Load() != 3 {
			t.Fatalf("expected 3 attempts, got %d", attempts.Load())
		}
	})
}

// ---------------------------------------------------------------------------
// Drill 5: Half-open single-flight
// SLO: While a probe request is in-flight in half-open state, all other
// requests must be rejected with ErrCircuitOpen.
// ---------------------------------------------------------------------------

func TestDrill_HalfOpenSingleFlight(t *testing.T) {
	const resetTimeout = 100 * time.Millisecond

	cb := platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
		Name:         "drill-halfopen",
		Threshold:    1,
		ResetTimeout: resetTimeout,
	})

	// Trip the circuit.
	_ = cb.Execute(func() error { return errors.New("fail") })
	if cb.GetState() != platform.StateOpen {
		t.Fatalf("expected open, got %s", cb.GetState().String())
	}

	// Wait for reset timeout.
	time.Sleep(resetTimeout + 20*time.Millisecond)

	// Start a slow probe request.
	probeDone := make(chan struct{})
	go func() {
		_ = cb.Execute(func() error {
			<-probeDone // Block until we release.
			return nil
		})
	}()

	// Give the probe goroutine time to start.
	time.Sleep(20 * time.Millisecond)

	// All other requests should be rejected.
	for i := 0; i < 5; i++ {
		err := cb.Execute(func() error { return nil })
		if !errors.Is(err, platform.ErrCircuitOpen) {
			t.Fatalf("request %d: expected ErrCircuitOpen during half-open probe, got %v", i, err)
		}
	}

	// Release the probe — circuit should close.
	close(probeDone)
	time.Sleep(10 * time.Millisecond)

	if cb.GetState() != platform.StateClosed {
		t.Fatalf("expected closed after probe success, got %s", cb.GetState().String())
	}
}
