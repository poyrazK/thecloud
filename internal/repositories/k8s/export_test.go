package k8s

import (
	"context"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) ExportEnsureClusterSecurityGroup(ctx context.Context, cluster *domain.Cluster) error {
	return p.ensureClusterSecurityGroup(ctx, cluster)
}

func (p *KubeadmProvisioner) ExportCreateNode(ctx context.Context, cluster *domain.Cluster, nameTag string, role domain.NodeRole) (*domain.Instance, error) {
	return p.createNode(ctx, cluster, nameTag, role)
}
