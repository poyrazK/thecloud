package k8s

import (
	"context"
	"fmt"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) ensureClusterSecurityGroup(ctx context.Context, cluster *domain.Cluster) error {
	sgName := fmt.Sprintf("sg-%s", cluster.Name)
	// Check if exists
	_, err := p.sgSvc.GetGroup(ctx, sgName, cluster.VpcID)
	if err == nil {
		return nil // Already exists
	}

	// Create
	sg, err := p.sgSvc.CreateGroup(ctx, cluster.VpcID, sgName, "Kubernetes cluster security group")
	if err != nil {
		return err
	}

	// Add Rules
	rules := []domain.SecurityRule{
		{Protocol: "tcp", PortMin: 6443, PortMax: 6443, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100},   // API Server
		{Protocol: "udp", PortMin: 4789, PortMax: 4789, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100},   // VXLAN
		{Protocol: "tcp", PortMin: 179, PortMax: 179, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100},     // BGP
		{Protocol: "tcp", PortMin: 10250, PortMax: 10250, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100}, // Kubelet
		{Protocol: "tcp", PortMin: 30000, PortMax: 32767, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100}, // NodePort TCP
		{Protocol: "udp", PortMin: 30000, PortMax: 32767, CIDR: AnyCIDR, Direction: domain.RuleIngress, Priority: 100}, // NodePort UDP
	}

	for _, r := range rules {
		_, _ = p.sgSvc.AddRule(ctx, sg.ID, r)
	}

	return nil
}

func (p *KubeadmProvisioner) applyBaseSecurity(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	p.logger.Info("applying base security manifests", "ip", masterIP, "isolation", cluster.NetworkIsolation)

	var securityManifests string
	if cluster.NetworkIsolation {
		// Apply a default-deny ingress policy for the default namespace
		securityManifests = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: default
spec:
  podSelector: {}
  policyTypes:
  - Ingress
`
	}
	if securityManifests != "" {
		manifestFile := "/tmp/base-security.yaml"
		_, _ = exec.Run(ctx, fmt.Sprintf("cat <<EOF > %s\n%s\nEOF", manifestFile, securityManifests))
		_, err = exec.Run(ctx, fmt.Sprintf("%s apply -f %s", kubectlBase, manifestFile))
		return err
	}

	return nil
}
