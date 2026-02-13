package k8s

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/mock"
)

type mockInstanceService struct{ mock.Mock }

func (m *mockInstanceService) LaunchInstance(ctx context.Context, params ports.LaunchParams) (*domain.Instance, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) Provision(ctx context.Context, job domain.ProvisionJob) error {
	return m.Called(ctx, job).Error(0)
}
func (m *mockInstanceService) LaunchInstanceWithOptions(ctx context.Context, opts ports.CreateInstanceOptions) (*domain.Instance, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) StartInstance(ctx context.Context, idOrName string) error { return nil }
func (m *mockInstanceService) StopInstance(ctx context.Context, idOrName string) error  { return nil }
func (m *mockInstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}
func (m *mockInstanceService) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	return "", nil
}
func (m *mockInstanceService) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	return nil, nil
}
func (m *mockInstanceService) GetConsoleURL(ctx context.Context, idOrName string) (string, error) {
	return "", nil
}
func (m *mockInstanceService) TerminateInstance(ctx context.Context, idOrName string) error {
	return nil
}
func (m *mockInstanceService) Exec(ctx context.Context, idOrName string, cmd []string) (string, error) {
	args := m.Called(ctx, idOrName, cmd)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *mockInstanceService) UpdateInstanceMetadata(ctx context.Context, id uuid.UUID, metadata, labels map[string]string) error {
	args := m.Called(ctx, id, metadata, labels)
	return args.Error(0)
}

type mockClusterRepo struct{ mock.Mock }

func (m *mockClusterRepo) Create(ctx context.Context, c *domain.Cluster) error { return nil }
func (m *mockClusterRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepo) ListAll(ctx context.Context) ([]*domain.Cluster, error) {
	return nil, nil
}
func (m *mockClusterRepo) Update(ctx context.Context, c *domain.Cluster) error      { return nil }
func (m *mockClusterRepo) Delete(ctx context.Context, id uuid.UUID) error           { return nil }
func (m *mockClusterRepo) AddNode(ctx context.Context, n *domain.ClusterNode) error { return nil }
func (m *mockClusterRepo) GetNodes(ctx context.Context, clusterID uuid.UUID) ([]*domain.ClusterNode, error) {
	args := m.Called(ctx, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ClusterNode), args.Error(1)
}
func (m *mockClusterRepo) DeleteNode(ctx context.Context, nodeID uuid.UUID) error      { return nil }
func (m *mockClusterRepo) UpdateNode(ctx context.Context, n *domain.ClusterNode) error { return nil }
