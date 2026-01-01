package ports

import (
	"context"

	"github.com/poyraz/cloud/internal/core/domain"
)

// DashboardService provides aggregated data for the web console.
type DashboardService interface {
	// GetSummary returns resource counts and overview metrics.
	GetSummary(ctx context.Context) (*domain.ResourceSummary, error)

	// GetRecentEvents returns the most recent audit events.
	GetRecentEvents(ctx context.Context, limit int) ([]*domain.Event, error)

	// GetStats returns the full dashboard statistics including metrics history.
	GetStats(ctx context.Context) (*domain.DashboardStats, error)
}
