// Package workers provides background worker implementations.
package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
)

const (
	provisionQueue      = "provision_queue"
	provisionGroup      = "provision_workers"
	provisionMaxWorkers = 20
	// How long a message can sit in PEL before another consumer reclaims it.
	// Must be shorter than provisionStaleThreshold (15m) to reclaim promptly,
	// but longer than the max expected job runtime (10m) to avoid stealing messages
	// from workers that are legitimately still processing.
	provisionReclaimMs = 12 * 60 * 1000 // 12 minutes
	provisionReclaimN  = 10
	// Stale threshold for idempotency ledger: if a "running" entry is older
	// than this, it is considered abandoned and can be reclaimed.
	provisionStaleThreshold = 15 * time.Minute
)

// ProvisionWorker processes instance provisioning tasks using a durable queue
// with at-least-once delivery. Jobs are acknowledged only after successful
// processing; crashed jobs are reclaimed by healthy peers. An execution ledger
// prevents duplicate processing of the same instance.
type ProvisionWorker struct {
	instSvc      *services.InstanceService
	taskQueue    ports.DurableTaskQueue
	ledger       ports.ExecutionLedger
	logger       *slog.Logger
	consumerName string
}

// NewProvisionWorker constructs a ProvisionWorker.
// If ledger is nil, idempotency checks are skipped.
func NewProvisionWorker(instSvc *services.InstanceService, taskQueue ports.DurableTaskQueue, ledger ports.ExecutionLedger, logger *slog.Logger) *ProvisionWorker {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "provision-worker"
	}
	return &ProvisionWorker{
		instSvc:      instSvc,
		taskQueue:    taskQueue,
		ledger:       ledger,
		logger:       logger,
		consumerName: hostname,
	}
}

func (w *ProvisionWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting provision worker",
		"consumer", w.consumerName,
		"concurrency", provisionMaxWorkers,
	)

	// Ensure consumer group exists.
	if err := w.taskQueue.EnsureGroup(ctx, provisionQueue, provisionGroup); err != nil {
		w.logger.Error("failed to ensure provision consumer group", "error", err)
		return
	}

	sem := make(chan struct{}, provisionMaxWorkers)

	// Start a background goroutine that periodically reclaims stale messages
	// from crashed consumers.
	go w.reclaimLoop(ctx, sem)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping provision worker")
			return
		default:
			msg, err := w.taskQueue.Receive(ctx, provisionQueue, provisionGroup, w.consumerName)
			if err != nil {
				w.logger.Error("failed to receive provision job", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if msg == nil {
				continue
			}

			var job domain.ProvisionJob
			if err := json.Unmarshal([]byte(msg.Payload), &job); err != nil {
				w.logger.Error("failed to unmarshal provision job",
					"error", err, "msg_id", msg.ID)
				// Ack poison messages so they don't block the queue.
				w.ackWithLog(ctx, msg.ID, "provision poison message")
				continue
			}

			w.logger.Info("processing provision job",
				"instance_id", job.InstanceID,
				"tenant_id", job.TenantID,
				"msg_id", msg.ID,
			)

			sem <- struct{}{} // acquire concurrency slot
			go func(m *ports.DurableMessage, j domain.ProvisionJob) {
				defer func() { <-sem }()
				w.processJob(ctx, m, j)
			}(msg, job)
		}
	}
}

func (w *ProvisionWorker) processJob(workerCtx context.Context, msg *ports.DurableMessage, job domain.ProvisionJob) {
	jobKey := fmt.Sprintf("provision:%s", job.InstanceID)

	// Idempotency check: skip if already completed or actively being processed.
	if w.ledger != nil {
		acquired, err := w.ledger.TryAcquire(workerCtx, jobKey, provisionStaleThreshold)
		if err != nil {
			w.logger.Error("execution ledger error",
				"instance_id", job.InstanceID, "msg_id", msg.ID, "error", err)
			// On ledger error, nack to retry later.
			w.nackWithLog(workerCtx, msg.ID, "ledger try_acquire failed")
			return
		}
		if !acquired {
			// Check if it's already finished or just being processed by someone else.
			status, _, _, getErr := w.ledger.GetStatus(workerCtx, jobKey)
			if getErr == nil && status == "completed" {
				w.logger.Info("skipping already completed provision job",
					"instance_id", job.InstanceID, "msg_id", msg.ID)
				w.ackWithLog(workerCtx, msg.ID, "provision already completed")
				return
			}
			w.logger.Info("provision job is currently being processed by another worker",
				"instance_id", job.InstanceID, "msg_id", msg.ID)
			return // Leave unacked for redelivery/wait.
		}
	}

	// Root context for background task with 10-minute safety timeout.
	ctx, cancel := context.WithTimeout(workerCtx, 10*time.Minute)
	defer cancel()

	// Inject User and Tenant IDs for repository access control.
	ctx = appcontext.WithUserID(ctx, job.UserID)
	ctx = appcontext.WithTenantID(ctx, job.TenantID)

	w.logger.Info("starting provision logic", "instance_id", job.InstanceID, "msg_id", msg.ID)
	if err := w.instSvc.Provision(ctx, job); err != nil {
		w.logger.Error("failed to provision instance",
			"instance_id", job.InstanceID,
			"msg_id", msg.ID,
			"error", err,
		)
		// Mark failed in the ledger so it can be retried.
		if w.ledger != nil {
			if ledgerErr := w.ledger.MarkFailed(workerCtx, jobKey, err.Error()); ledgerErr != nil {
				w.logger.Warn("failed to mark provision job failed in ledger",
					"instance_id", job.InstanceID, "msg_id", msg.ID, "error", ledgerErr)
			}
		}
		// Nack: leave message in PEL for reclaim/retry.
		w.nackWithLog(workerCtx, msg.ID, "provision failed")
		return
	}

	w.logger.Info("successfully provisioned instance",
		"instance_id", job.InstanceID,
		"msg_id", msg.ID,
	)

	// Mark completed in ledger (prevents duplicate execution).
	if w.ledger != nil {
		if ledgerErr := w.ledger.MarkComplete(workerCtx, jobKey, "ok"); ledgerErr != nil {
			w.logger.Warn("failed to mark provision job complete in ledger",
				"instance_id", job.InstanceID, "msg_id", msg.ID, "error", ledgerErr)
		}
	}

	// Acknowledge — message is permanently consumed.
	w.ackWithLog(workerCtx, msg.ID, "provision success")
}

// reclaimLoop periodically reclaims messages stuck in the PEL from crashed
// consumers and re-processes them.
func (w *ProvisionWorker) reclaimLoop(ctx context.Context, sem chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msgs, err := w.taskQueue.ReclaimStale(ctx, provisionQueue, provisionGroup, w.consumerName, provisionReclaimMs, provisionReclaimN)
			if err != nil {
				w.logger.Warn("provision reclaim error", "error", err)
				continue
			}
			for _, m := range msgs {
				var job domain.ProvisionJob
				if err := json.Unmarshal([]byte(m.Payload), &job); err != nil {
					w.logger.Error("failed to unmarshal reclaimed provision job",
						"msg_id", m.ID, "error", err)
					w.ackWithLog(ctx, m.ID, "reclaimed provision poison message")
					continue
				}
				w.logger.Info("reclaimed stale provision job",
					"instance_id", job.InstanceID, "msg_id", m.ID)

				m := m // capture loop variable
				sem <- struct{}{}
				go func() {
					defer func() { <-sem }()
					w.processJob(ctx, &m, job)
				}()
			}
		}
	}
}

func (w *ProvisionWorker) ackWithLog(ctx context.Context, messageID string, reason string) {
	if err := w.taskQueue.Ack(ctx, provisionQueue, provisionGroup, messageID); err != nil {
		w.logger.Warn("failed to ack provision job",
			"msg_id", messageID, "reason", reason, "error", err)
	}
}

func (w *ProvisionWorker) nackWithLog(ctx context.Context, messageID string, reason string) {
	if err := w.taskQueue.Nack(ctx, provisionQueue, provisionGroup, messageID); err != nil {
		w.logger.Warn("failed to nack provision job",
			"msg_id", messageID, "reason", reason, "error", err)
	}
}
