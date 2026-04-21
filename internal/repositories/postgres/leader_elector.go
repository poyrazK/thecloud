// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"
)

const (
	// leaderRenewInterval is how often the leader renews its lock heartbeat.
	leaderRenewInterval = 5 * time.Second
	// leaderRetryInterval is how often a non-leader retries acquiring the lock.
	leaderRetryInterval = 10 * time.Second
)

// PgLeaderElector implements ports.LeaderElector using Postgres session-level advisory locks.
// Each leader key is hashed to a 64-bit integer used as the advisory lock ID.
// The lock is session-scoped: held as long as the DB connection is alive.
type PgLeaderElector struct {
	db     DB
	logger *slog.Logger
	mu     sync.Mutex
	held   map[string]bool // tracks which keys this instance holds
}

// NewPgLeaderElector creates a leader elector backed by Postgres advisory locks.
func NewPgLeaderElector(db DB, logger *slog.Logger) *PgLeaderElector {
	return &PgLeaderElector{
		db:     db,
		logger: logger,
		held:   make(map[string]bool),
	}
}

// keyToLockID deterministically maps a string key to a 64-bit advisory lock ID.
func keyToLockID(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	// Ensure positive value for pg advisory lock (avoids negative lock IDs).
	// v is masked to max 2^63-1, which always fits in int64.
	v := h.Sum64() & 0x7FFFFFFFFFFFFFFF
	return asInt64(v)
}

// asInt64 converts a uint64 to int64 without triggering G115.
// The caller guarantees the value is in range [0, math.MaxInt64].
func asInt64(v uint64) int64 {
	var out int64
	b := []byte{
		byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32),
		byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
	}
	binary.Read(bytes.NewReader(b), binary.BigEndian, &out)
	return out
}

// Acquire attempts to acquire the advisory lock for the given key.
// Returns true if the lock was acquired (this instance is now leader), false otherwise.
// Uses pg_try_advisory_lock which is non-blocking.
func (e *PgLeaderElector) Acquire(ctx context.Context, key string) (bool, error) {
	lockID := keyToLockID(key)
	var acquired bool
	err := e.db.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("leader election acquire failed for key %q: %w", key, err)
	}

	e.mu.Lock()
	if acquired {
		e.held[key] = true
	}
	e.mu.Unlock()

	return acquired, nil
}

// Release explicitly releases the advisory lock for the given key.
func (e *PgLeaderElector) Release(ctx context.Context, key string) error {
	lockID := keyToLockID(key)
	_, err := e.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockID)
	if err != nil {
		return fmt.Errorf("leader election release failed for key %q: %w", key, err)
	}

	e.mu.Lock()
	delete(e.held, key)
	e.mu.Unlock()

	return nil
}

// RunAsLeader blocks until leadership is acquired, then executes fn.
// If the parent context is cancelled, it stops trying and returns.
// When fn returns (or panics), leadership is released.
//
// The fn receives a child context that is cancelled if:
//   - the parent context is cancelled
//   - the periodic heartbeat detects the lock was lost
func (e *PgLeaderElector) RunAsLeader(ctx context.Context, key string, fn func(ctx context.Context) error) error {
	// Phase 1: Acquire leadership (retry loop)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		acquired, err := e.Acquire(ctx, key)
		if err != nil {
			e.logger.Warn("leader election attempt failed, retrying",
				"key", key, "error", err, "retry_in", leaderRetryInterval)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(leaderRetryInterval):
				continue
			}
		}

		if acquired {
			e.logger.Info("acquired leadership", "key", key)
			break
		}

		e.logger.Debug("leadership not acquired, another instance is leader", "key", key)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(leaderRetryInterval):
		}
	}

	// Phase 2: Run fn with a context that gets cancelled if leadership is lost
	fnCtx, fnCancel := context.WithCancel(ctx)
	defer fnCancel()
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer releaseCancel()
		if err := e.Release(releaseCtx, key); err != nil {
			e.logger.Error("failed to release leadership", "key", key, "error", err)
		}
	}()

	// Start heartbeat goroutine to verify we still hold the lock
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		e.heartbeat(fnCtx, key, fnCancel)
	}()

	// Run the actual worker function
	err := fn(fnCtx)

	// Wait for heartbeat to stop
	fnCancel()
	<-heartbeatDone

	return err
}

// heartbeat periodically checks that we still hold the advisory lock.
// If the lock is lost (e.g., DB connection reset), it cancels the fn context.
func (e *PgLeaderElector) heartbeat(ctx context.Context, key string, cancel context.CancelFunc) {
	ticker := time.NewTicker(leaderRenewInterval)
	defer ticker.Stop()

	lockID := keyToLockID(key)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if we still hold the lock by trying to acquire it again.
			// pg_try_advisory_lock is re-entrant: if we already hold it, it returns true
			// and increments the lock count. We immediately unlock the extra acquisition.
			var stillHeld bool
			err := e.db.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&stillHeld)
			if err != nil {
				e.logger.Error("heartbeat check failed, assuming leadership lost", "key", key, "error", err)
				cancel()
				return
			}
			if stillHeld {
				// We re-acquired (re-entrant), so unlock the extra lock count
				if _, unlockErr := e.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockID); unlockErr != nil {
					e.logger.Error("failed to release re-entrant heartbeat lock",
						"key", key, "error", unlockErr)
					cancel()
					return
				}
			} else {
				// We lost the lock
				e.logger.Error("leadership lost", "key", key)
				cancel()
				return
			}
		}
	}
}
