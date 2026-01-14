// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// DashboardService provides business logic for aggregating system-wide metrics and resource summaries for the management console.
type DashboardService interface {
	// GetSummary returns an aggregated count of all provisioned cloud resources.
	GetSummary(ctx context.Context) (*domain.ResourceSummary, error)

	// GetRecentEvents retrieves the most recent system activities for the global activity feed.
	GetRecentEvents(ctx context.Context, limit int) ([]*domain.Event, error)

	// GetStats provides comprehensive dashboard data including real-time performance metrics and historical usage trends.
	GetStats(ctx context.Context) (*domain.DashboardStats, error)
}
