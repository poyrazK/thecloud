// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// AuditRepository defines the persistence layer for audit logs.
type AuditRepository interface {
	// Create persists a new audit log entry.
	Create(ctx context.Context, log *domain.AuditLog) error
	// ListByUserID retrieves the most recent audit logs for a specific user.
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error)
}

// AuditService provides business logic for recording and retrieving audit trails.
type AuditService interface {
	// Log generates and persists an audit entry based on user activity.
	Log(ctx context.Context, userID uuid.UUID, action, resourceType, resourceID string, details map[string]interface{}) error
	// ListLogs fetches a user's activity history.
	ListLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error)
}

// AuditLogger is a low-level interface for components that emit audit entries.
type AuditLogger interface {
	// Log emits a pre-constructed audit log entry.
	Log(ctx context.Context, entry *domain.AuditLog) error
}
