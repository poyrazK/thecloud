package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ClusterRepository defines the data access layer for Kubernetes clusters.
type ClusterRepository interface {
	Create(ctx context.Context, cluster *domain.Cluster) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error)
	Update(ctx context.Context, cluster *domain.Cluster) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Node operations
	AddNode(ctx context.Context, node *domain.ClusterNode) error
	GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error)
	DeleteNode(ctx context.Context, nodeID uuid.UUID) error
}

// ClusterService defines the business logic layer for Kubernetes clusters.
type ClusterService interface {
	CreateCluster(ctx context.Context, userID uuid.UUID, name string, vpcID uuid.UUID, version string, workers int) (*domain.Cluster, error)
	GetCluster(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	ListClusters(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error)
	DeleteCluster(ctx context.Context, id uuid.UUID) error
	GetKubeconfig(ctx context.Context, id uuid.UUID) (string, error)
}

// ClusterProvisioner defines the interface for bootstrapping a K8s cluster.
type ClusterProvisioner interface {
	Provision(ctx context.Context, cluster *domain.Cluster) error
	Deprovision(ctx context.Context, cluster *domain.Cluster) error
	GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error)
}
