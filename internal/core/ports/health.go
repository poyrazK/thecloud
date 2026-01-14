// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"time"
)

// HealthCheckResult encapsulates the aggregate health status of the system and its dependencies.
type HealthCheckResult struct {
	Status string            `json:"status"` // Overall status ("UP" or "DOWN")
	Checks map[string]string `json:"checks"` // Detailed status per component (e.g., "db": "OK", "redis": "FAIL")
	Time   time.Time         `json:"time"`   // High-resolution timestamp of when the check was performed
}

// HealthService provides business logic for performing system-wide health and readiness checks.
type HealthService interface {
	// Check aggregates health indicators from all core system components.
	Check(ctx context.Context) HealthCheckResult
}
