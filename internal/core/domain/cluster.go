// Package domain contains core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClusterStatus represents the current state of a Kubernetes cluster.
type ClusterStatus string

// Managed cluster states.
const (
	ClusterStatusPending      ClusterStatus = "pending"
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusRunning      ClusterStatus = "running"
	ClusterStatusUpgrading    ClusterStatus = "upgrading"
	ClusterStatusUpdating     ClusterStatus = "updating"
	ClusterStatusRepairing    ClusterStatus = "repairing"
	ClusterStatusFailed       ClusterStatus = "failed"
	ClusterStatusDeleting     ClusterStatus = "deleting"
)

// NodeRole represents the role of a node in a cluster.
type NodeRole string

// Cluster node roles.
const (
	NodeRoleControlPlane NodeRole = "control-plane"
	NodeRoleWorker       NodeRole = "worker"
)

// NodeGroup represents a pool of similar worker nodes in a cluster.
type NodeGroup struct {
	ID           uuid.UUID `json:"id"`
	ClusterID    uuid.UUID `json:"cluster_id"`
	Name         string    `json:"name"`
	InstanceType string    `json:"instance_type"`
	MinSize      int       `json:"min_size" example:"1"`
	MaxSize      int       `json:"max_size" example:"10"`
	CurrentSize  int       `json:"current_size" example:"3"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Cluster represents a managed Kubernetes cluster.
type Cluster struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	UserID          uuid.UUID     `json:"user_id"`
	TenantID        uuid.UUID     `json:"tenant_id"`
	VpcID           uuid.UUID     `json:"vpc_id"`
	Version         string        `json:"version"`
	ControlPlaneIPs []string      `json:"control_plane_ips"`
	WorkerCount     int           `json:"worker_count"` // Deprecated: use node_groups
	Status          ClusterStatus `json:"status"`
	// NodeGroups contains the node pools for this cluster.
	NodeGroups []NodeGroup `json:"node_groups,omitempty"`
	// Networking
	PodCIDR     string `json:"pod_cidr"`
	ServiceCIDR string `json:"service_cidr"`
	// Secrets (Encrypted)
	SSHPrivateKeyEncrypted string     `json:"-"`
	KubeconfigEncrypted    string     `json:"-"`
	JoinToken              string     `json:"-"`
	TokenExpiresAt         *time.Time `json:"-"`
	CACertHash             string     `json:"-"`

	NetworkIsolation   bool    `json:"network_isolation"`
	HAEnabled          bool    `json:"ha_enabled"`
	APIServerLBAddress *string `json:"api_server_lb_address,omitempty"`
	JobID              *string `json:"job_id,omitempty"`

	// Backup Policy
	BackupSchedule      string `json:"backup_schedule,omitempty" example:"0 0 * * *"`
	BackupRetentionDays int    `json:"backup_retention_days,omitempty" example:"7"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ClusterNode represents a node within a Kubernetes cluster.
type ClusterNode struct {
	ID            uuid.UUID  `json:"id"`
	ClusterID     uuid.UUID  `json:"cluster_id"`
	InstanceID    uuid.UUID  `json:"instance_id"`
	Role          NodeRole   `json:"role"`
	Status        string     `json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	JoinedAt      time.Time  `json:"joined_at"`
}
