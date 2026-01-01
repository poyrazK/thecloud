package ports

import (
	"context"

	"github.com/poyraz/cloud/internal/core/domain"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	List(ctx context.Context, limit int) ([]*domain.Event, error)
}

type EventService interface {
	RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error
	ListEvents(ctx context.Context, limit int) ([]*domain.Event, error)
}
