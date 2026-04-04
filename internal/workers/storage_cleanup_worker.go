// Package workers hosts background worker implementations.
package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	defaultCleanupInterval  = 1 * time.Hour
	defaultBatchSize        = 100
	defaultPendingUploadTTL = 1 * time.Hour
)

// StorageCleanupWorker periodically removes objects marked as deleted and orphaned pending uploads.
type StorageCleanupWorker struct {
	storageSvc       ports.StorageService
	logger           *slog.Logger
	interval         time.Duration
	batchSize        int
	pendingUploadTTL time.Duration
}

// NewStorageCleanupWorker constructs a StorageCleanupWorker.
func NewStorageCleanupWorker(storageSvc ports.StorageService, logger *slog.Logger) *StorageCleanupWorker {
	return &StorageCleanupWorker{
		storageSvc:       storageSvc,
		logger:           logger,
		interval:         defaultCleanupInterval,
		batchSize:        defaultBatchSize,
		pendingUploadTTL: defaultPendingUploadTTL,
	}
}

// SetTTL allows overriding the default TTL for pending uploads.
func (w *StorageCleanupWorker) SetTTL(ttl time.Duration) {
	w.pendingUploadTTL = ttl
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

	// 1. Cleanup soft-deleted objects
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

		w.logger.Debug("processed deleted cleanup batch", "count", count)
	}

	if totalDeleted > 0 {
		w.logger.Info("storage deleted cleanup completed", "total_deleted", totalDeleted)
	}

	// 2. Cleanup orphaned pending uploads (older than configurable TTL)
	totalPendingCleaned := 0
	for {
		count, err := w.storageSvc.CleanupPendingUploads(ctx, w.pendingUploadTTL, w.batchSize)
		if err != nil {
			w.logger.Error("failed to cleanup pending uploads", "error", err)
			break
		}

		totalPendingCleaned += count
		if count < w.batchSize {
			break
		}

		w.logger.Debug("processed pending cleanup batch", "count", count)
	}

	if totalPendingCleaned > 0 {
		w.logger.Info("storage pending cleanup completed", "total_cleaned", totalPendingCleaned)
	}
}
