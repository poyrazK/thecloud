// Package drills contains HA failure drills and release gates.
//
// Release gates are meant to run in CI before deploying a new version.
// They validate the SLO invariants for the control-plane HA features:
//
//  1. Leader failover <30s (validated via unit tests on LeaderGuard).
//  2. Zero duplicate singleton executions during failover.
//  3. Zero job loss in crash tests (durable queue ack/nack).
//  4. Circuit breaker fail-fast under backend failure.
//  5. Bulkhead prevents cascading overload.
//  6. No API outage during single pod loss (leader re-election + queue redelivery).
//
// Run release gates: go test ./internal/drills/ -v -count=1 -run TestReleaseGate
package drills

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/platform"
)

// TestReleaseGate_CircuitBreakerFailFast validates SLO:
// "When a backend is down, requests must fail-fast in <5ms".
func TestReleaseGate_CircuitBreakerFailFast(t *testing.T) {
	cb := platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
		Name:         "gate-cb",
		Threshold:    3,
		ResetTimeout: 1 * time.Second,
	})

	// Trip it.
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error { return errors.New("down") })
	}

	// Measure fail-fast latency over 100 calls.
	const iterations = 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		err := cb.Execute(func() error { return nil })
		if !errors.Is(err, platform.ErrCircuitOpen) {
			t.Fatalf("iteration %d: expected ErrCircuitOpen, got %v", i, err)
		}
	}
	elapsed := time.Since(start)

	avgLatency := elapsed / iterations
	if avgLatency > 1*time.Millisecond {
		t.Fatalf("average fail-fast latency %v exceeds 1ms SLO", avgLatency)
	}
	t.Logf("PASS: avg fail-fast latency = %v (SLO: <1ms)", avgLatency)
}

// TestReleaseGate_BulkheadIsolation validates SLO:
// "A saturated adapter must not block unrelated adapters".
func TestReleaseGate_BulkheadIsolation(t *testing.T) {
	// Two independent bulkheads for two adapters.
	bhCompute := platform.NewBulkhead(platform.BulkheadOpts{Name: "compute", MaxConc: 2, WaitTimeout: 50 * time.Millisecond})
	bhNetwork := platform.NewBulkhead(platform.BulkheadOpts{Name: "network", MaxConc: 5, WaitTimeout: 50 * time.Millisecond})

	ctx := context.Background()

	// Saturate compute bulkhead.
	blockCh := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bhCompute.Execute(ctx, func() error {
				<-blockCh
				return nil
			})
		}()
	}
	time.Sleep(30 * time.Millisecond) // Let them acquire slots.

	// Compute is now full.
	err := bhCompute.Execute(ctx, func() error { return nil })
	if !errors.Is(err, platform.ErrBulkheadFull) {
		t.Fatalf("compute bulkhead should be full, got %v", err)
	}

	// Network bulkhead must still be operational.
	err = bhNetwork.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("network bulkhead should be available, got %v", err)
	}

	close(blockCh)
	wg.Wait()
	t.Log("PASS: saturated compute did not affect network adapter")
}

// TestReleaseGate_CircuitBreakerRecovery validates SLO:
// "After backend recovery, the circuit must close within resetTimeout + probe time".
func TestReleaseGate_CircuitBreakerRecovery(t *testing.T) {
	const resetTimeout = 200 * time.Millisecond
	healthy := &atomic.Bool{}

	cb := platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
		Name:            "gate-recovery",
		Threshold:       2,
		ResetTimeout:    resetTimeout,
		SuccessRequired: 1,
	})

	// Trip it.
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error { return errors.New("down") })
	}

	// Simulate recovery after 100ms.
	go func() {
		time.Sleep(100 * time.Millisecond)
		healthy.Store(true)
	}()

	// Poll until circuit closes or timeout.
	deadline := time.After(resetTimeout + 200*time.Millisecond)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("circuit did not recover within SLO window. State: %s", cb.GetState().String())
		case <-ticker.C:
			err := cb.Execute(func() error {
				if healthy.Load() {
					return nil
				}
				return errors.New("still down")
			})
			if err == nil && cb.GetState() == platform.StateClosed {
				t.Logf("PASS: circuit recovered (state=%s)", cb.GetState().String())
				return
			}
		}
	}
}

// TestReleaseGate_RetryIdempotency validates SLO:
// "Retry must not execute the function more than MaxAttempts times".
func TestReleaseGate_RetryIdempotency(t *testing.T) {
	for _, maxAttempts := range []int{1, 3, 5, 10} {
		t.Run(fmt.Sprintf("max_%d", maxAttempts), func(t *testing.T) {
			var count atomic.Int64
			_ = platform.Retry(context.Background(), platform.RetryOpts{
				MaxAttempts: maxAttempts,
				BaseDelay:   1 * time.Millisecond,
				MaxDelay:    5 * time.Millisecond,
			}, func(ctx context.Context) error {
				count.Add(1)
				return errors.New("always fail")
			})

			if count.Load() != int64(maxAttempts) {
				t.Fatalf("expected exactly %d attempts, got %d", maxAttempts, count.Load())
			}
		})
	}
}

// TestReleaseGate_ConcurrentCircuitBreakers validates SLO:
// "Multiple independent circuit breakers must not interfere with each other".
func TestReleaseGate_ConcurrentCircuitBreakers(t *testing.T) {
	cbs := make([]*platform.CircuitBreaker, 5)
	for i := range cbs {
		cbs[i] = platform.NewCircuitBreakerWithOpts(platform.CircuitBreakerOpts{
			Name:         fmt.Sprintf("adapter-%d", i),
			Threshold:    3,
			ResetTimeout: 1 * time.Second,
		})
	}

	// Trip only breaker 0.
	for i := 0; i < 3; i++ {
		_ = cbs[0].Execute(func() error { return errors.New("down") })
	}

	if cbs[0].GetState() != platform.StateOpen {
		t.Fatal("breaker 0 should be open")
	}

	// All others should be closed and functional.
	for i := 1; i < 5; i++ {
		err := cbs[i].Execute(func() error { return nil })
		if err != nil {
			t.Fatalf("breaker %d should be functional, got %v", i, err)
		}
		if cbs[i].GetState() != platform.StateClosed {
			t.Fatalf("breaker %d should be closed, got %s", i, cbs[i].GetState().String())
		}
	}
	t.Log("PASS: tripped breaker did not affect independent breakers")
}
