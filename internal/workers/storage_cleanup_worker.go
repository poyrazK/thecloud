// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// StorageCleanupWorker periodically removes objects marked as deleted.
type StorageCleanupWorker struct {
	storageSvc ports.StorageService
	logger     *slog.Logger
	interval   time.Duration
	batchSize  int
}

// NewStorageCleanupWorker constructs a StorageCleanupWorker.
func NewStorageCleanupWorker(storageSvc ports.StorageService, logger *slog.Logger) *StorageCleanupWorker {
	return &StorageCleanupWorker{
		storageSvc: storageSvc,
		logger:     logger,
		interval:   1 * time.Hour,
		batchSize:  100,
	}
}

func (w *StorageCleanupWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("storage cleanup worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("storage cleanup worker stopping")
			return
		case <-ticker.C:
			w.cleanup(ctx)
		}
	}
}

func (w *StorageCleanupWorker) cleanup(ctx context.Context) {
	w.logger.Info("running storage cleanup")

	totalDeleted := 0
	for {
		count, err := w.storageSvc.CleanupDeleted(ctx, w.batchSize)
		if err != nil {
			w.logger.Error("failed to cleanup deleted objects", "error", err)
			break
		}

		totalDeleted += count
		if count < w.batchSize {
			break
		}

		w.logger.Debug("processed cleanup batch", "count", count)
	}

	if totalDeleted > 0 {
		w.logger.Info("storage cleanup completed", "total_deleted", totalDeleted)
	}
}
