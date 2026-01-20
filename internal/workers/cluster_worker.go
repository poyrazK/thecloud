// Package workers provides background worker implementations for various cloud tasks.
package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ClusterWorker handles background tasks for Kubernetes cluster lifecycle management.
type ClusterWorker struct {
	repo        ports.ClusterRepository
	provisioner ports.ClusterProvisioner
	taskQueue   ports.TaskQueue
	logger      *slog.Logger
}

// NewClusterWorker creates a new ClusterWorker.
func NewClusterWorker(repo ports.ClusterRepository, provisioner ports.ClusterProvisioner, taskQueue ports.TaskQueue, logger *slog.Logger) *ClusterWorker {
	return &ClusterWorker{
		repo:        repo,
		provisioner: provisioner,
		taskQueue:   taskQueue,
		logger:      logger,
	}
}

const (
	queuePollBackoff    = 1 * time.Second
	maxConcurrentClusts = 10
)

func (w *ClusterWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting cluster worker", "concurrency", maxConcurrentClusts)

	sem := make(chan struct{}, maxConcurrentClusts)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping cluster worker")
			return
		default:
			msg, err := w.taskQueue.Dequeue(ctx, "k8s_jobs")
			if err != nil {
				w.logger.Error("failed to dequeue cluster job", "error", err)
				time.Sleep(queuePollBackoff)
				continue
			}
			if msg == "" {
				time.Sleep(queuePollBackoff)
				continue
			}

			var job domain.ClusterJob
			if err := json.Unmarshal([]byte(msg), &job); err != nil {
				w.logger.Error("failed to unmarshal cluster job", "error", err)
				continue
			}

			w.logger.Info("processing cluster job", "cluster_id", job.ClusterID, "type", job.Type)

			sem <- struct{}{}
			go func() {
				defer func() { <-sem }()
				w.processJob(job)
			}()
		}
	}
}

func (w *ClusterWorker) processJob(job domain.ClusterJob) {
	// Root context for background task
	ctx := appcontext.WithUserID(context.Background(), job.UserID)

	cluster, err := w.repo.GetByID(ctx, job.ClusterID)
	if err != nil {
		w.logger.Error("failed to fetch cluster for job", "cluster_id", job.ClusterID, "error", err)
		return
	}
	if cluster == nil {
		w.logger.Error("cluster not found for job", "cluster_id", job.ClusterID)
		return
	}

	switch job.Type {
	case domain.ClusterJobProvision:
		w.handleProvision(ctx, cluster)
	case domain.ClusterJobDeprovision:
		w.handleDeprovision(ctx, cluster)
	case domain.ClusterJobUpgrade:
		w.handleUpgrade(ctx, cluster, job.Version)
	}
}

func (w *ClusterWorker) handleProvision(ctx context.Context, cluster *domain.Cluster) {
	cluster.Status = domain.ClusterStatusProvisioning
	cluster.UpdatedAt = time.Now()
	_ = w.repo.Update(ctx, cluster)

	if err := w.provisioner.Provision(ctx, cluster); err != nil {
		w.logger.Error("provisioning failed", "cluster_id", cluster.ID, "error", err)
		cluster.Status = domain.ClusterStatusFailed
	} else {
		w.logger.Info("provisioning succeeded", "cluster_id", cluster.ID)
		cluster.Status = domain.ClusterStatusRunning
	}

	cluster.UpdatedAt = time.Now()
	cluster.JobID = nil // Clear job ID
	_ = w.repo.Update(ctx, cluster)
}

func (w *ClusterWorker) handleDeprovision(ctx context.Context, cluster *domain.Cluster) {
	cluster.Status = domain.ClusterStatusDeleting
	cluster.UpdatedAt = time.Now()
	_ = w.repo.Update(ctx, cluster)

	if err := w.provisioner.Deprovision(ctx, cluster); err != nil {
		w.logger.Error("deprovisioning failed", "cluster_id", cluster.ID, "error", err)
		// We might still mark it as failed or just leave it
	} else {
		w.logger.Info("deprovisioning succeeded", "cluster_id", cluster.ID)
		_ = w.repo.Delete(ctx, cluster.ID)
		return
	}

	cluster.UpdatedAt = time.Now()
	cluster.JobID = nil
	_ = w.repo.Update(ctx, cluster)
}

func (w *ClusterWorker) handleUpgrade(ctx context.Context, cluster *domain.Cluster, version string) {
	cluster.Status = domain.ClusterStatusUpgrading
	cluster.UpdatedAt = time.Now()
	_ = w.repo.Update(ctx, cluster)

	if err := w.provisioner.Upgrade(ctx, cluster, version); err != nil {
		w.logger.Error("upgrade failed", "cluster_id", cluster.ID, "error", err)
		cluster.Status = domain.ClusterStatusRunning // Revert to running if failed
	} else {
		w.logger.Info("upgrade succeeded", "cluster_id", cluster.ID)
		cluster.Status = domain.ClusterStatusRunning
		cluster.Version = version
	}

	cluster.UpdatedAt = time.Now()
	cluster.JobID = nil
	_ = w.repo.Update(ctx, cluster)
}
