// Package domain defines core business entities.
package domain

import (
	"github.com/google/uuid"
)

// ProvisionJob represents an asynchronous task to initialize and configure a new compute instance.
type ProvisionJob struct {
	InstanceID uuid.UUID          `json:"instance_id"`
	UserID     uuid.UUID          `json:"user_id"`
	TenantID   uuid.UUID          `json:"tenant_id"`
	Volumes    []VolumeAttachment `json:"volumes"` // List of storage volumes to attach during initialization
}

// ClusterJobType describes the kind of cluster lifecycle operation.
type ClusterJobType string

const (
	// ClusterJobProvision creates a new Kubernetes cluster.
	ClusterJobProvision ClusterJobType = "provision"
	// ClusterJobDeprovision deletes an existing Kubernetes cluster.
	ClusterJobDeprovision ClusterJobType = "deprovision"
	// ClusterJobUpgrade upgrades a cluster to a new version.
	ClusterJobUpgrade ClusterJobType = "upgrade"
)

// ClusterJob represents a background task for Kubernetes cluster operations.
type ClusterJob struct {
	ClusterID uuid.UUID      `json:"cluster_id"`
	UserID    uuid.UUID      `json:"user_id"`
	Type      ClusterJobType `json:"type"`
	Version   string         `json:"version,omitempty"` // For upgrades
}
