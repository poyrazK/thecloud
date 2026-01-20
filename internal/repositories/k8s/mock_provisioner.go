package k8s

import (
	"context"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// MockProvisioner is a simulation provisioner for testing.
type MockProvisioner struct{}

func NewMockProvisioner() *MockProvisioner {
	return &MockProvisioner{}
}

func (p *MockProvisioner) Provision(ctx context.Context, cluster *domain.Cluster) error {
	// Simulate work
	time.Sleep(2 * time.Second)
	cluster.ControlPlaneIPs = []string{"10.0.0.10"}
	cluster.Kubeconfig = "apiVersion: v1\nclusters:\n- cluster:\n    server: https://10.0.0.10:6443\n  name: mock-cluster"
	return nil
}

func (p *MockProvisioner) Deprovision(ctx context.Context, cluster *domain.Cluster) error {
	// Simulate work
	time.Sleep(1 * time.Second)
	return nil
}

func (p *MockProvisioner) GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error) {
	return domain.ClusterStatusRunning, nil
}

func (p *MockProvisioner) Repair(ctx context.Context, cluster *domain.Cluster) error {
	return nil
}

func (p *MockProvisioner) Scale(ctx context.Context, cluster *domain.Cluster) error {
	return nil
}

func (p *MockProvisioner) GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error) {
	return "mock-kubeconfig", nil
}

func (p *MockProvisioner) GetHealth(ctx context.Context, cluster *domain.Cluster) (*ports.ClusterHealth, error) {
	return &ports.ClusterHealth{Status: cluster.Status, APIServer: true, NodesTotal: cluster.WorkerCount + 1, NodesReady: cluster.WorkerCount + 1}, nil
}

func (p *MockProvisioner) Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error {
	return nil
}

func (p *MockProvisioner) RotateSecrets(ctx context.Context, cluster *domain.Cluster) error {
	return nil
}

func (p *MockProvisioner) CreateBackup(ctx context.Context, cluster *domain.Cluster) error {
	return nil
}

func (p *MockProvisioner) Restore(ctx context.Context, cluster *domain.Cluster, backupPath string) error {
	return nil
}
