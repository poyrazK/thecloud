// Package workers provides background worker implementations.
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
	"github.com/poyrazk/thecloud/internal/core/services"
)

// ProvisionWorker processes instance provisioning tasks.
type ProvisionWorker struct {
	instSvc   *services.InstanceService
	taskQueue ports.TaskQueue
	logger    *slog.Logger
}

// NewProvisionWorker constructs a ProvisionWorker.
func NewProvisionWorker(instSvc *services.InstanceService, taskQueue ports.TaskQueue, logger *slog.Logger) *ProvisionWorker {
	return &ProvisionWorker{
		instSvc:   instSvc,
		taskQueue: taskQueue,
		logger:    logger,
	}
}

func (w *ProvisionWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting provision worker")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping provision worker")
			return
		default:
			// Dequeue task
			msg, err := w.taskQueue.Dequeue(ctx, "provision_queue")
			if err != nil {
				// redis.Nil or other error
				time.Sleep(1 * time.Second)
				continue
			}

			if msg == "" {
				continue
			}

			var job domain.ProvisionJob
			if err := json.Unmarshal([]byte(msg), &job); err != nil {
				w.logger.Error("failed to unmarshal provision job", "error", err)
				continue
			}

			w.logger.Info("processing provision job", "instance_id", job.InstanceID, "tenant_id", job.TenantID)

			// Process job concurrently to handle high throughput in load tests
			go w.processJob(job)
		}
	}
}

func (w *ProvisionWorker) processJob(job domain.ProvisionJob) {
	// Root context for background task with 10-minute safety timeout
	// We use context.Background() because the worker lifecycle context shouldn't necessarily cancel active provisioning unless the app is shutting down
	baseCtx := context.Background()
	ctx, cancel := context.WithTimeout(baseCtx, 10*time.Minute)
	defer cancel()

	// Inject User and Tenant IDs for repository access control
	ctx = appcontext.WithUserID(ctx, job.UserID)
	ctx = appcontext.WithTenantID(ctx, job.TenantID)

	w.logger.Info("starting provision logic", "instance_id", job.InstanceID)
	if err := w.instSvc.Provision(ctx, job.InstanceID, job.Volumes); err != nil {
		w.logger.Error("failed to provision instance", "instance_id", job.InstanceID, "error", err)
	} else {
		w.logger.Info("successfully provisioned instance", "instance_id", job.InstanceID)
	}
}
