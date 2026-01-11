package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

type AccountingWorker struct {
	accountingSvc ports.AccountingService
	logger        *slog.Logger
	interval      time.Duration
}

func NewAccountingWorker(accountingSvc ports.AccountingService, logger *slog.Logger) *AccountingWorker {
	return &AccountingWorker{
		accountingSvc: accountingSvc,
		logger:        logger,
		interval:      1 * time.Minute, // For development, check every minute instead of hour
	}
}

func (w *AccountingWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	w.logger.Info("starting accounting worker", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping accounting worker")
			return
		case <-ticker.C:
			w.logger.Debug("processing billing cycle")
			if err := w.accountingSvc.ProcessHourlyBilling(ctx); err != nil {
				w.logger.Error("failed to process billing", "error", err)
			}
		}
	}
}
