package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClusterStatus represents the current state of a Kubernetes cluster.
type ClusterStatus string

const (
	ClusterStatusPending      ClusterStatus = "pending"
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusRunning      ClusterStatus = "running"
	ClusterStatusUpgrading    ClusterStatus = "upgrading"
	ClusterStatusFailed       ClusterStatus = "failed"
	ClusterStatusDeleting     ClusterStatus = "deleting"
)

// NodeRole represents the role of a node in a cluster.
type NodeRole string

const (
	NodeRoleControlPlane NodeRole = "control-plane"
	NodeRoleWorker       NodeRole = "worker"
)

// Cluster represents a managed Kubernetes cluster.
type Cluster struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	UserID          uuid.UUID     `json:"user_id"`
	VpcID           uuid.UUID     `json:"vpc_id"`
	Version         string        `json:"version"`
	ControlPlaneIPs []string      `json:"control_plane_ips"`
	WorkerCount     int           `json:"worker_count"`
	Status          ClusterStatus `json:"status"`
	Kubeconfig      string        `json:"kubeconfig,omitempty"` // Encrypted
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
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
