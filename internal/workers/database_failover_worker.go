// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	defaultDatabaseFailoverInterval = 30 * time.Second
	databaseCheckTimeout            = 2 * time.Second
)

// DatabaseFailoverWorker monitors managed database primaries and performs automatic failover to replicas.
type DatabaseFailoverWorker struct {
	dbSvc  ports.DatabaseService
	repo   ports.DatabaseRepository
	logger *slog.Logger

	interval time.Duration
}

// NewDatabaseFailoverWorker constructs a DatabaseFailoverWorker.
func NewDatabaseFailoverWorker(dbSvc ports.DatabaseService, repo ports.DatabaseRepository, logger *slog.Logger) *DatabaseFailoverWorker {
	return &DatabaseFailoverWorker{
		dbSvc:    dbSvc,
		repo:     repo,
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

func (w *DatabaseFailoverWorker) isHealthy(_ context.Context, db *domain.Database) bool {
	// Simple TCP check for the mapped port on localhost (since it's a simulator)
	address := fmt.Sprintf("localhost:%d", db.Port)
	conn, err := net.DialTimeout("tcp", address, databaseCheckTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
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

	// Select the first available replica for promotion
	replica := replicas[0]
	w.logger.Info("promoting replica to primary", "replica_id", replica.ID, "primary_id", primary.ID)

	if err := w.dbSvc.PromoteToPrimary(ctx, replica.ID); err != nil {
		w.logger.Error("failed to promote replica", "replica_id", replica.ID, "error", err)
		return
	}

	w.logger.Info("successfully promoted replica to primary", "replica_id", replica.ID)
}
