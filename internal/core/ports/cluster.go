// Package ports defines interfaces for adapters and services.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ClusterHealth represents the operational status of a K8s cluster.
type ClusterHealth struct {
	Status     domain.ClusterStatus `json:"status"`
	APIServer  bool                 `json:"api_server"`
	NodesTotal int                  `json:"nodes_total"`
	NodesReady int                  `json:"nodes_ready"`
	Message    string               `json:"message,omitempty"`
}

// ClusterRepository defines the data access layer for Kubernetes clusters.
type ClusterRepository interface {
	Create(ctx context.Context, cluster *domain.Cluster) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error)
	ListAll(ctx context.Context) ([]*domain.Cluster, error)
	Update(ctx context.Context, cluster *domain.Cluster) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Node operations
	AddNode(ctx context.Context, node *domain.ClusterNode) error
	GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error)
	UpdateNode(ctx context.Context, node *domain.ClusterNode) error
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error
}

// CreateClusterParams defines the options for cluster creation.
type CreateClusterParams struct {
	UserID           uuid.UUID
	Name             string
	VpcID            uuid.UUID
	Version          string
	Workers          int
	NetworkIsolation bool
	HAEnabled        bool
}

// ClusterService defines the business logic layer for Kubernetes clusters.
type ClusterService interface {
	CreateCluster(ctx context.Context, params CreateClusterParams) (*domain.Cluster, error)
	GetCluster(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	ListClusters(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error)
	DeleteCluster(ctx context.Context, id uuid.UUID) error
	GetKubeconfig(ctx context.Context, id uuid.UUID, role string) (string, error)
	RepairCluster(ctx context.Context, id uuid.UUID) error
	ScaleCluster(ctx context.Context, id uuid.UUID, workers int) error
	GetClusterHealth(ctx context.Context, id uuid.UUID) (*ClusterHealth, error)
	UpgradeCluster(ctx context.Context, id uuid.UUID, version string) error
	RotateSecrets(ctx context.Context, id uuid.UUID) error
	CreateBackup(ctx context.Context, id uuid.UUID) error
	RestoreBackup(ctx context.Context, id uuid.UUID, backupPath string) error
}

// ClusterProvisioner defines the interface for bootstrapping a K8s cluster.
type ClusterProvisioner interface {
	Provision(ctx context.Context, cluster *domain.Cluster) error
	Deprovision(ctx context.Context, cluster *domain.Cluster) error
	GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error)
	Repair(ctx context.Context, cluster *domain.Cluster) error
	Scale(ctx context.Context, cluster *domain.Cluster) error
	GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error)
	GetHealth(ctx context.Context, cluster *domain.Cluster) (*ClusterHealth, error)
	Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error
	RotateSecrets(ctx context.Context, cluster *domain.Cluster) error
	CreateBackup(ctx context.Context, cluster *domain.Cluster) error
	Restore(ctx context.Context, cluster *domain.Cluster, backupPath string) error
}
