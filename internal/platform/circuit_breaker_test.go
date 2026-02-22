package platform

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/require"
	"time"

	"github.com/stretchr/testify/assert"
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

		// This should be allowed (half-open state implicitly)
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
