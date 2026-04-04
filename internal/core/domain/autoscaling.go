// Package domain contains the core domain models for the cloud platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// System-wide hard limits (Security requirement)
const (
	MaxInstancesHardLimit  = 20 // Maximum instances per scaling group to prevent resource exhaustion
	MaxScalingGroupsPerVPC = 5  // Maximum scaling groups per VPC
	MinCooldownSeconds     = 60 // Minimum cooldown to prevent rapid thrashing
)

// ScalingGroupStatus represents the lifecycle state of a scaling group.
type ScalingGroupStatus string

const (
	// ScalingGroupStatusActive indicates the group is functioning normally.
	ScalingGroupStatusActive ScalingGroupStatus = "ACTIVE"
	// ScalingGroupStatusUpdating indicates the group configuration is being modified.
	ScalingGroupStatusUpdating ScalingGroupStatus = "UPDATING"
	// ScalingGroupStatusDeleting indicates the group is being torn down.
	ScalingGroupStatusDeleting ScalingGroupStatus = "DELETING"
	// ScalingGroupStatusDeleted indicates the group has been removed.
	ScalingGroupStatusDeleted ScalingGroupStatus = "DELETED"
)

// ScalingGroup represents a collection of instances that scale horizontally.
// It maintains a desired number of instances between MinInstances and MaxInstances
// based on defined scaling policies.
type ScalingGroup struct {
	ID             uuid.UUID          `json:"id"`
	UserID         uuid.UUID          `json:"user_id"`
	TenantID       uuid.UUID          `json:"tenant_id"`
	IdempotencyKey string             `json:"idempotency_key,omitempty"` // For safe retries
	Name           string             `json:"name"`
	VpcID          uuid.UUID          `json:"vpc_id"`
	LoadBalancerID *uuid.UUID         `json:"load_balancer_id,omitempty"` // Optional LB integration
	Image          string             `json:"image"`                      // Instance image (e.g. "nginx")
	InstanceType   string             `json:"instance_type"`              // NEW: Configuration type
	Ports          string             `json:"ports,omitempty"`            // Ports exposed by instances
	MinInstances   int                `json:"min_instances"`              // Floor for scaling
	MaxInstances   int                `json:"max_instances"`              // Ceiling for scaling
	DesiredCount   int                `json:"desired_count"`              // Target number of instances
	CurrentCount   int                `json:"current_count"`              // Actual number of instances
	Status         ScalingGroupStatus `json:"status"`
	FailureCount   int                `json:"failure_count"` // Consecutive failure tracker
	LastFailureAt  *time.Time         `json:"last_failure_at,omitempty"`
	Version        int                `json:"version"` // Optimistic locking
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ScalingPolicy defines rules for automatic scaling actions.
// It uses metrics (CPU, Memory) to trigger scale-out or scale-in events.
type ScalingPolicy struct {
	ID             uuid.UUID  `json:"id"`
	ScalingGroupID uuid.UUID  `json:"scaling_group_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Name           string     `json:"name"`
	MetricType     string     `json:"metric_type"`    // "cpu" | "memory"
	TargetValue    float64    `json:"target_value"`   // Threshold (e.g. 80.0 for 80%)
	ScaleOutStep   int        `json:"scale_out_step"` // Instances to add
	ScaleInStep    int        `json:"scale_in_step"`  // Instances to remove
	CooldownSec    int        `json:"cooldown_sec"`   // Wait time after scaling
	LastScaledAt   *time.Time `json:"last_scaled_at,omitempty"`
}

// ScalingGroupInstance maps an instance to its parent scaling group.
type ScalingGroupInstance struct {
	ScalingGroupID uuid.UUID `json:"scaling_group_id"`
	InstanceID     uuid.UUID `json:"instance_id"`
	JoinedAt       time.Time `json:"joined_at"`
}
