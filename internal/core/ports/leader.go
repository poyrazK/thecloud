// Package ports defines service and repository interfaces.
package ports

import (
	"context"
)

// LeaderElector provides distributed leader election for singleton controllers.
// Only one instance across all replicas should hold leadership for a given key at any time.
type LeaderElector interface {
	// Acquire attempts to become the leader for the given key.
	// It returns true if leadership was acquired, false otherwise.
	// The leadership is held until Release is called or the context is cancelled.
	Acquire(ctx context.Context, key string) (bool, error)

	// Release relinquishes leadership for the given key.
	Release(ctx context.Context, key string) error

	// RunAsLeader blocks until leadership is acquired for the given key, then calls fn.
	// If leadership is lost, fn's context is cancelled. If fn returns, leadership is released.
	// This is the primary entrypoint for singleton workers.
	RunAsLeader(ctx context.Context, key string, fn func(ctx context.Context) error) error
}
