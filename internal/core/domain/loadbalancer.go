// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// LBStatus represents the lifecycle state of a load balancer.
type LBStatus string

const (
	// LBStatusCreating indicates the load balancer is being provisioned.
	LBStatusCreating LBStatus = "CREATING"
	// LBStatusActive indicates the load balancer is routing traffic.
	LBStatusActive LBStatus = "ACTIVE"
	// LBStatusDraining indicates connections are being finished before shutdown.
	LBStatusDraining LBStatus = "DRAINING"
	// LBStatusDeleted indicates the load balancer has been removed.
	LBStatusDeleted LBStatus = "DELETED"
)

// LoadBalancer distributes incoming traffic across multiple targets.
// It ensures high availability and scalability of applications.
type LoadBalancer struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	Name           string    `json:"name"`
	VpcID          uuid.UUID `json:"vpc_id"`
	Port           int       `json:"port"`      // Listener port (e.g. 80)
	Algorithm      string    `json:"algorithm"` // "round-robin" | "least-conn"
	Status         LBStatus  `json:"status"`
	Version        int       `json:"version"` // Optimistic locking
	CreatedAt      time.Time `json:"created_at"`
}

// LBTarget represents a backend instance that receives traffic.
type LBTarget struct {
	ID         uuid.UUID `json:"id"`
	LBID       uuid.UUID `json:"lb_id"`
	InstanceID uuid.UUID `json:"instance_id"`
	Port       int       `json:"port"`
	Weight     int       `json:"weight"` // Traffic share for weighted algorithms
	Health     string    `json:"health"` // "healthy" | "unhealthy" | "unknown"
}

// HealthCheckConfig defines how the load balancer checks target health.
type HealthCheckConfig struct {
	Path               string `json:"path"`                // HTTP path (e.g. "/health")
	IntervalSeconds    int    `json:"interval_seconds"`    // Frequency of checks
	TimeoutSeconds     int    `json:"timeout_seconds"`     // Max time to wait
	HealthyThreshold   int    `json:"healthy_threshold"`   // Successes needed for "healthy"
	UnhealthyThreshold int    `json:"unhealthy_threshold"` // Failures needed for "unhealthy"
}
