// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/poyrazk/thecloud/internal/repositories/postgres"
)

// ReplicaMonitor periodically checks read replica health and updates DualDB.
type ReplicaMonitor struct {
	db        *postgres.DualDB
	replica   postgres.DB
	interval  time.Duration
	logger    *slog.Logger
	isHealthy atomic.Bool
}

// NewReplicaMonitor creates a new worker to monitor database replicas.
func NewReplicaMonitor(db *postgres.DualDB, replica postgres.DB, logger *slog.Logger) *ReplicaMonitor {
	m := &ReplicaMonitor{
		db:       db,
		replica:  replica,
		interval: 30 * time.Second,
		logger:   logger.With("worker", "replica_monitor"),
	}
	m.isHealthy.Store(replica != nil)
	return m
}

// Run starts the monitoring loop.
func (m *ReplicaMonitor) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if m.replica == nil {
		m.logger.Info("no replica configured, skipping monitor")
		return
	}

	m.logger.Info("starting replica monitor", "interval", m.interval)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Initial check
	m.check(ctx)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("stopping replica monitor")
			return
		case <-ticker.C:
			m.check(ctx)
		}
	}
}

func (m *ReplicaMonitor) check(ctx context.Context) {
	err := m.replica.Ping(ctx)
	wasHealthy := m.isHealthy.Swap(err == nil)

	if err != nil {
		if wasHealthy {
			m.logger.Error("replica database heartbeat failed, failing over to primary", "error", err)
		}
		m.db.SetReplicaHealthy(false)
	} else {
		if !wasHealthy {
			m.logger.Info("replica database recovered, enabling read routing")
		}
		m.db.SetReplicaHealthy(true)
	}
}

// IsHealthy returns the current reported health of the replica.
func (m *ReplicaMonitor) IsHealthy() bool {
	return m.isHealthy.Load()
}
