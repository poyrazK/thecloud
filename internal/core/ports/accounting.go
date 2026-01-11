package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type AccountingService interface {
	TrackUsage(ctx context.Context, record domain.UsageRecord) error
	GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.BillSummary, error)
	ListUsage(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error)
	ProcessHourlyBilling(ctx context.Context) error
}

type AccountingRepository interface {
	CreateRecord(ctx context.Context, record domain.UsageRecord) error
	GetUsageSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (map[domain.ResourceType]float64, error)
	ListRecords(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error)
}
