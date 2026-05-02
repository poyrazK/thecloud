// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	defaultDatabaseFailoverInterval = 30 * time.Second
	databaseCheckTimeout            = 2 * time.Second
	maxAcceptableLagSeconds         = 5
)

// DatabaseFailoverWorker monitors managed database primaries and performs automatic failover to replicas.
type DatabaseFailoverWorker struct {
	dbSvc   ports.DatabaseService
	repo    ports.DatabaseRepository
	compute ports.ComputeBackend
	logger  *slog.Logger

	interval time.Duration
}

// NewDatabaseFailoverWorker constructs a DatabaseFailoverWorker.
func NewDatabaseFailoverWorker(dbSvc ports.DatabaseService, repo ports.DatabaseRepository, compute ports.ComputeBackend, logger *slog.Logger) *DatabaseFailoverWorker {
	return &DatabaseFailoverWorker{
		dbSvc:    dbSvc,
		repo:     repo,
		compute:  compute,
		logger:   logger.With("worker", "database_failover"),
		interval: defaultDatabaseFailoverInterval,
	}
}

// Run starts the failover monitoring loop.
func (w *DatabaseFailoverWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting database failover worker", "interval", w.interval)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping database failover worker")
			return
		case <-ticker.C:
			w.checkDatabases(ctx)
		}
	}
}

func (w *DatabaseFailoverWorker) checkDatabases(ctx context.Context) {
	dbs, err := w.repo.List(ctx)
	if err != nil {
		w.logger.Error("failed to list databases for failover check", "error", err)
		return
	}

	for _, db := range dbs {
		if db.Role != domain.RolePrimary || db.Status != domain.DatabaseStatusRunning {
			continue
		}

		if !w.isHealthy(ctx, db) {
			w.logger.Warn("detected primary database failure, initiating failover", "id", db.ID, "name", db.Name)
			w.handleFailover(ctx, db)
		}
	}
}

func (w *DatabaseFailoverWorker) isHealthy(ctx context.Context, db *domain.Database) bool {
	if db.Port == 0 {
		return true
	}
	address := fmt.Sprintf("127.0.0.1:%d", db.Port)
	dialer := &net.Dialer{Timeout: databaseCheckTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}
	if err := conn.Close(); err != nil {
		w.logger.Debug("failed to close connection", "address", address, "error", err)
	}
	return true
}

func (w *DatabaseFailoverWorker) handleFailover(ctx context.Context, primary *domain.Database) {
	replicas, err := w.repo.ListReplicas(ctx, primary.ID)
	if err != nil {
		w.logger.Error("failed to list replicas for failover", "primary_id", primary.ID, "error", err)
		return
	}

	if len(replicas) == 0 {
		w.logger.Error("failover failed: no replicas available for primary", "primary_id", primary.ID)
		return
	}

	replica, err := w.selectBestReplica(ctx, replicas)
	if err != nil {
		w.logger.Error("no healthy replica found for failover", "primary_id", primary.ID, "error", err)
		return
	}

	w.logger.Info("promoting replica to primary", "replica_id", replica.ID, "primary_id", primary.ID)

	if err := w.dbSvc.PromoteToPrimary(ctx, replica.ID); err != nil {
		w.logger.Error("failed to promote replica", "replica_id", replica.ID, "error", err)
		return
	}

	w.logger.Info("successfully promoted replica to primary", "replica_id", replica.ID)
}

// selectBestReplica selects the healthiest replica with the lowest replication lag.
func (w *DatabaseFailoverWorker) selectBestReplica(ctx context.Context, replicas []*domain.Database) (*domain.Database, error) {
	var best *domain.Database
	bestLag := int(^uint(0) >> 1) // max int

	for _, replica := range replicas {
		if replica.Status != domain.DatabaseStatusRunning {
			continue
		}
		lag, healthy := w.checkReplicationStatus(ctx, replica)
		if !healthy {
			continue
		}
		if lag < bestLag {
			bestLag = lag
			best = replica
		}
	}

	if best == nil {
		return nil, errors.New("no healthy replica found")
	}
	return best, nil
}

// checkReplicationStatus checks the replication lag on a replica.
// Returns lag in seconds and whether the replica is healthy enough for promotion.
func (w *DatabaseFailoverWorker) checkReplicationStatus(ctx context.Context, replica *domain.Database) (lagSeconds int, healthy bool) {
	if replica.Engine != domain.EnginePostgres {
		// For non-PostgreSQL engines, do a simple TCP check
		return 0, w.isHealthy(ctx, replica)
	}

	// Query standby-side recovery metrics for PostgreSQL replica.
	// pg_stat_replication is primary-side only; use standby-side metrics.
	// - pg_is_in_recovery() = true on standbys, false on primaries
	// - pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() means fully caught up
	// - pg_last_xact_replay_timestamp() measures actual replay lag
	query := `SELECT CASE
    WHEN NOT pg_is_in_recovery() THEN 2147483647
    WHEN pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() THEN 0
    WHEN pg_last_xact_replay_timestamp() IS NULL THEN 2147483647
    ELSE EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp()))::INTEGER
END AS lag_seconds;`

	execCtx, cancel := context.WithTimeout(ctx, databaseCheckTimeout)
	defer cancel()

	output, err := w.compute.Exec(execCtx, replica.ContainerID, []string{"psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query})
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			w.logger.Debug("replication query timed out", "replica_id", replica.ID)
		} else {
			w.logger.Debug("failed to query replication status", "replica_id", replica.ID, "error", err)
		}
		return 0, false
	}

	// Parse lag from output (output is a single integer value)
	var lag int
	if _, parseErr := fmt.Sscanf(strings.TrimSpace(output), "%d", &lag); parseErr != nil {
		w.logger.Debug("failed to parse replication lag", "replica_id", replica.ID, "output", output)
		return 0, false
	}

	return lag, lag <= maxAcceptableLagSeconds
}
