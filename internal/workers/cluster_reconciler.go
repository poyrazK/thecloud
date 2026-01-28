// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ClusterReconciler periodically checks all clusters and repairs them if needed.
type ClusterReconciler struct {
	repo        ports.ClusterRepository
	provisioner ports.ClusterProvisioner
	interval    time.Duration
	logger      *slog.Logger
}

// NewClusterReconciler creates a new ClusterReconciler.
func NewClusterReconciler(repo ports.ClusterRepository, provisioner ports.ClusterProvisioner, logger *slog.Logger) *ClusterReconciler {
	return &ClusterReconciler{
		repo:        repo,
		provisioner: provisioner,
		interval:    5 * time.Minute,
		logger:      logger.With("worker", "cluster_reconciler"),
	}
}

// Run starts the reconciliation loop.
func (r *ClusterReconciler) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	r.logger.Info("starting cluster reconciler", "interval", r.interval)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Initial run
	r.reconcile(ctx)

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("stopping cluster reconciler")
			return
		case <-ticker.C:
			r.reconcile(ctx)
		}
	}
}

func (r *ClusterReconciler) reconcile(ctx context.Context) {
	clusters, err := r.repo.ListAll(ctx)
	if err != nil {
		r.logger.Error("failed to list clusters for reconciliation", "error", err)
		return
	}

	for _, cluster := range clusters {
		// Only reconcile active clusters
		if cluster.Status != domain.ClusterStatusRunning {
			continue
		}

		r.logger.Debug("checking cluster health", "cluster_id", cluster.ID, "name", cluster.Name)
		health, err := r.provisioner.GetHealth(ctx, cluster)
		if err != nil {
			r.logger.Error("failed to get cluster health", "cluster_id", cluster.ID, "error", err)
			continue
		}

		// Simple reconciliation logic: if API is down or nodes are not ready, attempt a repair
		if !health.APIServer || health.NodesReady < health.NodesTotal {
			r.logger.Warn("cluster unhealthy, initiating repair",
				"cluster_id", cluster.ID,
				"api_server", health.APIServer,
				"nodes_ready", health.NodesReady,
				"nodes_total", health.NodesTotal,
			)

			if err := r.provisioner.Repair(ctx, cluster); err != nil {
				r.logger.Error("failed to repair cluster", "cluster_id", cluster.ID, "error", err)
			} else {
				r.logger.Info("cluster repair initiated successfully", "cluster_id", cluster.ID)
			}
		}
	}
}
