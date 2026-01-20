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

// Cluster represents a managed Kubernetes cluster.
type Cluster struct {
	ID                 uuid.UUID     `json:"id"`
	Name               string        `json:"name"`
	UserID             uuid.UUID     `json:"user_id"`
	VpcID              uuid.UUID     `json:"vpc_id"`
	Version            string        `json:"version"`
	ControlPlaneIPs    []string      `json:"control_plane_ips"`
	WorkerCount        int           `json:"worker_count"`
	Status             ClusterStatus `json:"status"`
	SSHKey             string        `json:"-"` // Base64 encoded private key
	Kubeconfig         string        `json:"-"` // Encrypted
	NetworkIsolation   bool          `json:"network_isolation"`
	HAEnabled          bool          `json:"ha_enabled"`
	APIServerLBAddress *string       `json:"api_server_lb_address,omitempty"`
	JobID              *string       `json:"job_id,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// ClusterNode represents a node within a Kubernetes cluster.
type ClusterNode struct {
	ID         uuid.UUID `json:"id"`
	ClusterID  uuid.UUID `json:"cluster_id"`
	InstanceID uuid.UUID `json:"instance_id"`
	Role       NodeRole  `json:"role"`
	Status     string    `json:"status"`
	JoinedAt   time.Time `json:"joined_at"`
}
