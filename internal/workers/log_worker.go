package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// LogWorker handles background tasks for the CloudLogs service.
type LogWorker struct {
	logSvc ports.LogService
	logger *slog.Logger
	interval time.Duration
	retentionDays int
}

// NewLogWorker constructs a LogWorker.
func NewLogWorker(logSvc ports.LogService, logger *slog.Logger) *LogWorker {
	return &LogWorker{
		logSvc:        logSvc,
		logger:        logger,
		interval:      24 * time.Hour, // Run retention once a day
		retentionDays: 30,            // Default 30 days retention
	}
}

func (w *LogWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("log worker started")

	// Use a system context for background tasks to avoid RBAC issues
	systemCtx := appcontext.WithUserID(ctx, appcontext.SystemUserID) 

	// Initial run
	if err := w.logSvc.RunRetentionPolicy(systemCtx, w.retentionDays); err != nil {
		w.logger.Error("failed to run initial log retention policy", "error", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("log worker stopping")
			return
		case <-ticker.C:
			if err := w.logSvc.RunRetentionPolicy(systemCtx, w.retentionDays); err != nil {
				w.logger.Error("failed to run log retention policy", "error", err)
			}
		}
	}
}
