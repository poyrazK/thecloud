package ports

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// LogRepository defines the interface for persisting and retrieving log entries.
type LogRepository interface {
	// Create persists a batch of log entries.
	Create(ctx context.Context, entries []*domain.LogEntry) error
	
	// List retrieves log entries based on the provided query.
	List(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error)
	
	// DeleteByAge removes logs older than the specified duration (for retention).
	DeleteByAge(ctx context.Context, days int) error
}

// LogService defines the business logic for managing platform logs.
type LogService interface {
	// IngestLogs processes and persists logs for a resource.
	IngestLogs(ctx context.Context, entries []*domain.LogEntry) error
	
	// SearchLogs searches logs with filtering and pagination.
	SearchLogs(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error)
	
	// RunRetentionPolicy performs periodic log cleanup.
	RunRetentionPolicy(ctx context.Context, days int) error
}
