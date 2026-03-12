package platform

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("Success Path", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)
		err := cb.Execute(func() error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("Trip Circuit", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		// First failure
		err := cb.Execute(func() error {
			return errors.New("fail")
		})
		require.Error(t, err)
		assert.Equal(t, StateClosed, cb.GetState())

		// Second failure - trips circuit
		err = cb.Execute(func() error {
			return errors.New("fail")
		})
		require.Error(t, err)
		assert.Equal(t, StateOpen, cb.GetState())

		// Subsequent call returns ErrCircuitOpen
		err = cb.Execute(func() error {
			return nil
		})
		assert.Equal(t, ErrCircuitOpen, err)
	})

	t.Run("Reset After Timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 50*time.Millisecond)

		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
		assert.Equal(t, StateOpen, cb.GetState())

		time.Sleep(100 * time.Millisecond)

		// This should be allowed (half-open state)
		err := cb.Execute(func() error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("Half-Open Failure Retrips", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 50*time.Millisecond)

		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
		assert.Equal(t, StateOpen, cb.GetState())

		time.Sleep(100 * time.Millisecond)

		// This should be allowed but fail, retripping immediately
		err := cb.Execute(func() error {
			return errors.New("fail")
		})
		require.Error(t, err)
		assert.NotEqual(t, ErrCircuitOpen, err)
		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("Manual Reset", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 1*time.Hour)
		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
		assert.Equal(t, StateOpen, cb.GetState())

		cb.Reset()
		assert.Equal(t, StateClosed, cb.GetState())
	})
}

func TestCircuitBreakerHalfOpenSingleFlight(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)

	// Trip the circuit.
	_ = cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.GetState())

	time.Sleep(100 * time.Millisecond)

	// First call goes through as the half-open probe. Use a channel to
	// hold the probe in-flight while we test the second call.
	probeStarted := make(chan struct{})
	probeDone := make(chan struct{})

	go func() {
		_ = cb.Execute(func() error {
			close(probeStarted)
			<-probeDone // block until test releases
			return nil
		})
	}()

	<-probeStarted // wait for probe to be in-flight

	// Second concurrent call should be rejected while probe is in flight.
	err := cb.Execute(func() error { return nil })
	assert.Equal(t, ErrCircuitOpen, err, "second request should be blocked while half-open probe is in flight")

	close(probeDone) // release the probe
	time.Sleep(10 * time.Millisecond)

	// After probe succeeds, circuit should be closed.
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreakerOnStateChange(t *testing.T) {
	var mu sync.Mutex
	transitions := make([]struct{ from, to State }, 0)

	cb := NewCircuitBreakerWithOpts(CircuitBreakerOpts{
		Name:         "test-cb",
		Threshold:    1,
		ResetTimeout: 50 * time.Millisecond,
		OnStateChange: func(name string, from, to State) {
			mu.Lock()
			transitions = append(transitions, struct{ from, to State }{from, to})
			mu.Unlock()
		},
	})

	// Trip it.
	_ = cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(20 * time.Millisecond) // let async callback fire

	mu.Lock()
	require.Len(t, transitions, 1)
	assert.Equal(t, StateClosed, transitions[0].from)
	assert.Equal(t, StateOpen, transitions[0].to)
	mu.Unlock()

	// Wait for reset timeout, then succeed to close.
	time.Sleep(100 * time.Millisecond)
	err := cb.Execute(func() error { return nil })
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	// Should have: closed->open, open->half-open, half-open->closed
	require.Len(t, transitions, 3)
	assert.Equal(t, StateOpen, transitions[1].from)
	assert.Equal(t, StateHalfOpen, transitions[1].to)
	assert.Equal(t, StateHalfOpen, transitions[2].from)
	assert.Equal(t, StateClosed, transitions[2].to)
	mu.Unlock()
}

func TestCircuitBreakerWithOpts(t *testing.T) {
	cb := NewCircuitBreakerWithOpts(CircuitBreakerOpts{
		Name:            "compute",
		Threshold:       3,
		ResetTimeout:    1 * time.Second,
		SuccessRequired: 2,
	})

	assert.Equal(t, "compute", cb.Name())
	assert.Equal(t, StateClosed, cb.GetState())

	// Trip it with 3 failures.
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error { return errors.New("fail") })
	}
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreakerSuccessRequired(t *testing.T) {
	cb := NewCircuitBreakerWithOpts(CircuitBreakerOpts{
		Threshold:       1,
		ResetTimeout:    50 * time.Millisecond,
		SuccessRequired: 2,
	})

	// Trip it.
	_ = cb.Execute(func() error { return errors.New("fail") })
	assert.Equal(t, StateOpen, cb.GetState())

	time.Sleep(100 * time.Millisecond)

	// First success should move to half-open but not closed.
	err := cb.Execute(func() error { return nil })
	require.NoError(t, err)
	// Still half-open because we need 2 successes.
	assert.Equal(t, StateHalfOpen, cb.GetState())

	// Second success should close.
	err = cb.Execute(func() error { return nil })
	require.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestStateString(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "unknown(99)", State(99).String())
}
