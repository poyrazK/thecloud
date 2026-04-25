package workers

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockLeaderElector implements ports.LeaderElector for testing.
type mockLeaderElector struct {
	acquireResult bool
	acquireErr    error
	releaseErr    error
	acquireCount  atomic.Int32
	releaseCount  atomic.Int32

	// When set, RunAsLeader immediately calls fn if acquireResult is true
	runAsLeaderFn func(ctx context.Context, key string, fn func(ctx context.Context) error) error
}

func (m *mockLeaderElector) Acquire(ctx context.Context, key string) (bool, error) {
	m.acquireCount.Add(1)
	return m.acquireResult, m.acquireErr
}

func (m *mockLeaderElector) Release(ctx context.Context, key string) error {
	m.releaseCount.Add(1)
	return m.releaseErr
}

func (m *mockLeaderElector) RunAsLeader(ctx context.Context, key string, fn func(ctx context.Context) error) error {
	if m.runAsLeaderFn != nil {
		return m.runAsLeaderFn(ctx, key, fn)
	}
	// Default: acquire leadership and run fn
	if m.acquireResult {
		return fn(ctx)
	}
	// Not leader, block until context cancelled
	<-ctx.Done()
	return ctx.Err()
}

// mockRunner records whether Run was called and blocks until context is done.
type mockRunner struct {
	runCalled atomic.Int32
	runCtx    context.Context
}

func (r *mockRunner) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	r.runCalled.Add(1)
	r.runCtx = ctx
	<-ctx.Done()
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLeaderGuardRunsInnerWorkerWhenLeader(t *testing.T) {
	elector := &mockLeaderElector{acquireResult: true}
	inner := &mockRunner{}
	guard := NewLeaderGuard(elector, "test:worker", inner, newTestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go guard.Run(ctx, wg)

	// Wait a bit for the inner worker to start
	time.Sleep(100 * time.Millisecond)

	if inner.runCalled.Load() == 0 {
		t.Fatal("expected inner worker to be started when leader")
	}

	cancel()
	wg.Wait()
}

func TestLeaderGuardDoesNotRunWhenNotLeader(t *testing.T) {
	elector := &mockLeaderElector{
		acquireResult: false,
		runAsLeaderFn: func(ctx context.Context, key string, fn func(ctx context.Context) error) error {
			// Simulate never becoming leader — block until cancelled
			<-ctx.Done()
			return ctx.Err()
		},
	}
	inner := &mockRunner{}
	guard := NewLeaderGuard(elector, "test:worker", inner, newTestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go guard.Run(ctx, wg)

	wg.Wait()

	if inner.runCalled.Load() != 0 {
		t.Fatal("expected inner worker NOT to be started when not leader")
	}
}

func TestLeaderGuardRestartsAfterLeadershipLoss(t *testing.T) {
	callCount := atomic.Int32{}

	elector := &mockLeaderElector{
		runAsLeaderFn: func(ctx context.Context, key string, fn func(ctx context.Context) error) error {
			n := callCount.Add(1)
			if n <= 2 {
				// Simulate short leadership then loss
				fnCtx, fnCancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer fnCancel()
				return fn(fnCtx)
			}
			// Third time: block until parent context cancelled
			<-ctx.Done()
			return ctx.Err()
		},
	}

	inner := &mockRunner{}
	// Override mockRunner to not block
	countingRunner := &countingMockRunner{}
	guard := NewLeaderGuard(elector, "test:worker", countingRunner, newTestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go guard.Run(ctx, wg)

	wg.Wait()
	_ = inner // unused, countingRunner is used instead

	runs := countingRunner.runCalled.Load()
	if runs < 2 {
		t.Fatalf("expected inner worker to be restarted at least 2 times after leadership loss, got %d", runs)
	}
}

// countingMockRunner counts Run calls but returns quickly when context is done.
type countingMockRunner struct {
	runCalled atomic.Int32
}

func (r *countingMockRunner) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	r.runCalled.Add(1)
	<-ctx.Done()
}

func TestLeaderGuardShutsDownCleanly(t *testing.T) {
	elector := &mockLeaderElector{acquireResult: true}
	inner := &mockRunner{}
	guard := NewLeaderGuard(elector, "test:worker", inner, newTestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go guard.Run(ctx, wg)

	// Let it start
	time.Sleep(50 * time.Millisecond)

	// Cancel and wait for clean shutdown
	cancel()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success — clean shutdown
	case <-time.After(2 * time.Second):
		t.Fatal("leader guard did not shut down within 2s")
	}
}
