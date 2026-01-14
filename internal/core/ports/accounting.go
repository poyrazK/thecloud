// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// AccountingService defines the business logic for tracking resource usage and billing.
type AccountingService interface {
	// TrackUsage persists a single resource consumption event.
	TrackUsage(ctx context.Context, record domain.UsageRecord) error
	// GetSummary aggregates usage into a billable summary for a specific user and time range.
	GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.BillSummary, error)
	// ListUsage retrieves detailed usage records for a user within a time frame.
	ListUsage(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error)
	// ProcessHourlyBilling triggers system-wide calculations for the current billing cycle.
	ProcessHourlyBilling(ctx context.Context) error
}

// AccountingRepository handles the persistence of usage records and billing data.
type AccountingRepository interface {
	// CreateRecord saves a new usage record to the database.
	CreateRecord(ctx context.Context, record domain.UsageRecord) error
	// GetUsageSummary calculates aggregated quantities per resource type for a user.
	GetUsageSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (map[domain.ResourceType]float64, error)
	// ListRecords fetches raw usage data from storage.
	ListRecords(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]domain.UsageRecord, error)
}
