// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// EventRepository manages the persistent storage of system-wide events.
type EventRepository interface {
	// Create saves a new system event record.
	Create(ctx context.Context, event *domain.Event) error
	// List retrieves a specified number of recent system events.
	List(ctx context.Context, limit int) ([]*domain.Event, error)
}

// EventService provides business logic for recording and querying system activities.
type EventService interface {
	// RecordEvent creates an event entry with associated metadata.
	RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error
	// ListEvents returns a slice of recent events for observability purposes.
	ListEvents(ctx context.Context, limit int) ([]*domain.Event, error)
}
