package services

import (
	"context"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// Checkable defines a dependency that can be pinged for health checks.
type Checkable interface {
	Ping(ctx context.Context) error
}

// HealthServiceImpl aggregates system checks across dependencies.
type HealthServiceImpl struct {
	db      Checkable
	compute ports.ComputeBackend
}

// NewHealthServiceImpl constructs a health service with its dependencies.
func NewHealthServiceImpl(db Checkable, compute ports.ComputeBackend) *HealthServiceImpl {
	return &HealthServiceImpl{
		db:      db,
		compute: compute,
	}
}

func (s *HealthServiceImpl) Check(ctx context.Context) ports.HealthCheckResult {
	checks := make(map[string]string)
	overall := "UP"

	// Check DB
	if err := s.db.Ping(ctx); err != nil {
		checks["database"] = "DISCONNECTED: " + err.Error()
		overall = "DEGRADED"
	} else {
		checks["database"] = "CONNECTED"
	}

	// Check Docker
	if err := s.compute.Ping(ctx); err != nil {
		checks["docker"] = "DISCONNECTED: " + err.Error()
		overall = "DEGRADED"
	} else {
		checks["docker"] = "CONNECTED"
	}

	return ports.HealthCheckResult{
		Status: overall,
		Checks: checks,
		Time:   time.Now(),
	}
}
