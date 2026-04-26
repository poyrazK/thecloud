// Package workers provides background worker implementations.
package workers

import (
	"context"
	"log/slog"
	"sync"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// LeaderGuard wraps a worker that implements the Run(context.Context, *sync.WaitGroup)
// interface and ensures it only runs on the pod that holds leadership for its key.
//
// When leadership is not held, the worker is paused. If leadership is lost mid-run,
// the worker's context is cancelled, causing it to stop. It will restart if
// leadership is re-acquired.
type LeaderGuard struct {
	elector ports.LeaderElector
	key     string
	inner   runner
	logger  *slog.Logger
}

// runner is the interface all workers implement.
type runner interface {
	Run(context.Context, *sync.WaitGroup)
}

// NewLeaderGuard creates a LeaderGuard that protects the given worker with leader election.
// The key should be unique per worker type (e.g., "worker:lb", "worker:cron").
func NewLeaderGuard(elector ports.LeaderElector, key string, inner runner, logger *slog.Logger) *LeaderGuard {
	return &LeaderGuard{
		elector: elector,
		key:     key,
		inner:   inner,
		logger:  logger,
	}
}

// Run implements the runner interface. It participates in leader election and only
// runs the inner worker when this instance is the leader. If leadership is lost,
// the inner worker is stopped. If leadership is re-acquired, the inner worker restarts.
func (g *LeaderGuard) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if ctx.Err() != nil {
			return
		}

		g.logger.Info("attempting to acquire leadership", "key", g.key)

		err := g.elector.RunAsLeader(ctx, g.key, func(leaderCtx context.Context) error {
			g.logger.Info("running as leader", "key", g.key)

			// Create an inner WaitGroup for the wrapped worker
			innerWG := &sync.WaitGroup{}
			innerWG.Add(1)
			go g.inner.Run(leaderCtx, innerWG)

			// Wait for the inner worker to finish (either normally or due to context cancellation)
			innerWG.Wait()

			g.logger.Info("inner worker stopped", "key", g.key)
			return nil
		})

		if err != nil {
			if ctx.Err() != nil {
				// Parent context cancelled — clean shutdown
				g.logger.Info("leader guard shutting down", "key", g.key)
				return
			}
			g.logger.Error("leader election error, will retry", "key", g.key, "error", err)
		}

		// If we reach here, we either lost leadership or RunAsLeader returned.
		// Loop back to try to re-acquire leadership.
		if ctx.Err() != nil {
			return
		}
		g.logger.Info("leadership lost or released, retrying", "key", g.key)
	}
}
