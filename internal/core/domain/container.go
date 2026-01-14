// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// DeploymentStatus represents the current operational state of a container deployment.
type DeploymentStatus string

const (
	// DeploymentStatusScaling indicates the deployment is adjusting its replica count.
	DeploymentStatusScaling DeploymentStatus = "SCALING"
	// DeploymentStatusReady indicates all requested replicas are running and healthy.
	DeploymentStatusReady DeploymentStatus = "READY"
	// DeploymentStatusDegraded indicates some replicas are failing or unreachable.
	DeploymentStatusDegraded DeploymentStatus = "DEGRADED"
	// DeploymentStatusDeleting indicates the deployment is being removed.
	DeploymentStatusDeleting DeploymentStatus = "DELETING"
)

// Deployment represents a managed set of identical container replicas (CaaS).
type Deployment struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"user_id"`
	Name         string           `json:"name"`          // Unique name for the deployment
	Image        string           `json:"image"`         // Container image (e.g., "redis:alpine")
	Replicas     int              `json:"replicas"`      // Desired number of replicas
	CurrentCount int              `json:"current_count"` // Actual number of running replicas
	Ports        string           `json:"ports"`         // Exposed ports (e.g., "80:8080")
	Status       DeploymentStatus `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// DeploymentContainer links a specific container instance to its parent deployment.
type DeploymentContainer struct {
	ID           uuid.UUID `json:"id"`
	DeploymentID uuid.UUID `json:"deployment_id"`
	InstanceID   uuid.UUID `json:"instance_id"` // Reference to the underlying compute instance
}
