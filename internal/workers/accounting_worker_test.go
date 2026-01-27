package workers

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type countingAccountingService struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (t *countingAccountingService) ProcessHourlyBilling(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls++
	return t.err
}

func (t *countingAccountingService) Calls() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.calls
}

func (t *countingAccountingService) TrackUsage(ctx context.Context, record domain.UsageRecord) error {
	return nil
}

func (t *countingAccountingService) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.BillSummary, error) {
	return nil, nil
}

func (t *countingAccountingService) ListUsage(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error) {
	return nil, nil
}

func TestAccountingWorkerRun(t *testing.T) {
	fakeSvc := &countingAccountingService{}
	worker := &AccountingWorker{
		accountingSvc: fakeSvc,
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		interval:      10 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)

	time.Sleep(35 * time.Millisecond)
	cancel()
	wg.Wait()

	if got := fakeSvc.Calls(); got < 2 {
		t.Fatalf("expected at least 2 billing runs, got %d", got)
	}
}

func TestAccountingWorkerRunLogsErrors(t *testing.T) {
	svc := &countingAccountingService{err: errors.New("boom")}
	worker := &AccountingWorker{
		accountingSvc: svc,
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		interval:      5 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)

	time.Sleep(20 * time.Millisecond)
	cancel()
	wg.Wait()

	if got := svc.Calls(); got < 1 {
		t.Fatalf("expected billing to run despite errors, got %d", got)
	}
}
