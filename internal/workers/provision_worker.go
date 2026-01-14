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

			w.logger.Debug("processing provision job", "instance_id", job.InstanceID)

			// Process job concurrently to handle high throughput in load tests
			go func(job domain.ProvisionJob) {
				// We use a new context with UserID for authorization
				ctx := appcontext.WithUserID(context.Background(), job.UserID)
				if err := w.instSvc.Provision(ctx, job.InstanceID, job.Volumes); err != nil {
					w.logger.Error("failed to provision instance", "instance_id", job.InstanceID, "error", err)
				} else {
					w.logger.Debug("successfully provisioned instance", "instance_id", job.InstanceID)
				}
			}(job)
		}
	}
}
