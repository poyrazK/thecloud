// Package services implements core business workflows.
package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// Checkable defines a dependency that can be pinged for health checks.
type Checkable interface {
	Ping(ctx context.Context) error
}

// DatabaseStatusProvider defines an optional interface for databases that can report detailed status.
type DatabaseStatusProvider interface {
	GetStatus(ctx context.Context) map[string]string
}

// HealthServiceImpl aggregates system checks across dependencies.
type HealthServiceImpl struct {
	db      Checkable
	compute ports.ComputeBackend
	cluster ports.ClusterService
}

// NewHealthServiceImpl constructs a health service with its dependencies.
func NewHealthServiceImpl(db Checkable, compute ports.ComputeBackend, cluster ports.ClusterService) *HealthServiceImpl {
	return &HealthServiceImpl{
		db:      db,
		compute: compute,
		cluster: cluster,
	}
}

func (s *HealthServiceImpl) Check(ctx context.Context) ports.HealthCheckResult {
	checks := make(map[string]string)
	overall := "UP"

	s.checkDatabase(ctx, checks, &overall)
	s.checkCompute(ctx, checks, &overall)
	s.checkClusters(ctx, checks, &overall)

	return ports.HealthCheckResult{
		Status: overall,
		Checks: checks,
		Time:   time.Now(),
	}
}

func (s *HealthServiceImpl) checkDatabase(ctx context.Context, checks map[string]string, overall *string) {
	// Check Primary
	if err := s.db.Ping(ctx); err != nil {
		checks["database_primary"] = "DISCONNECTED: " + err.Error()
		*overall = "DEGRADED"
	} else {
		checks["database_primary"] = "CONNECTED"
	}

	// Check Replica if provider
	if provider, ok := s.db.(DatabaseStatusProvider); ok {
		dbStats := provider.GetStatus(ctx)
		for k, v := range dbStats {
			checks[k] = v
			if v != "CONNECTED" && v != "HEALTHY" && k == "database_replica" {
				if *overall == "UP" {
					*overall = "DEGRADED"
				}
			}
		}
	}
}

func (s *HealthServiceImpl) checkCompute(ctx context.Context, checks map[string]string, overall *string) {
	if err := s.compute.Ping(ctx); err != nil {
		checks["docker"] = "DISCONNECTED: " + err.Error()
		*overall = "DEGRADED"
	} else {
		checks["docker"] = "CONNECTED"
	}
}

func (s *HealthServiceImpl) checkClusters(ctx context.Context, checks map[string]string, overall *string) {
	if s.cluster == nil {
		return
	}

	_, err := s.cluster.ListClusters(appcontext.WithInternalCall(ctx), uuid.Nil)
	if err != nil {
		checks["kubernetes_service"] = "DEGRADED: " + err.Error()
		*overall = "DEGRADED"
	} else {
		checks["kubernetes_service"] = "OK"
	}
}
