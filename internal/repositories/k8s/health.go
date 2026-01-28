// Package k8s implements Kubernetes provisioning adapters.
package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

func (p *KubeadmProvisioner) GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error) {
	return cluster.Status, nil
}

func (p *KubeadmProvisioner) Repair(ctx context.Context, cluster *domain.Cluster) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	p.logger.Info("repairing cluster components", "cluster_id", cluster.ID, "ip", masterIP)

	// 1. Re-apply CNI
	if err := p.installCNI(ctx, cluster, masterIP); err != nil {
		return err
	}

	// 2. Re-apply kube-proxy patch
	if err := p.patchKubeProxy(ctx, cluster, masterIP); err != nil {
		return err
	}

	// 3. Re-apply base security
	if err := p.applyBaseSecurity(ctx, cluster, masterIP); err != nil {
		return err
	}

	// 4. Reconcile node count to ensure any missing/failed nodes are replaced
	return p.Scale(ctx, cluster)
}

func (p *KubeadmProvisioner) Scale(ctx context.Context, cluster *domain.Cluster) error {
	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err != nil {
		return err
	}

	currentCount := 0
	var workers []*domain.ClusterNode
	for _, n := range nodes {
		if n.Role == domain.NodeRoleWorker {
			currentCount++
			workers = append(workers, n)
		}
	}

	if cluster.WorkerCount > currentCount {
		return p.scaleUp(ctx, cluster, currentCount)
	} else if cluster.WorkerCount < currentCount {
		return p.scaleDown(ctx, cluster, workers)
	}

	return nil
}

func (p *KubeadmProvisioner) scaleUp(ctx context.Context, cluster *domain.Cluster, currentCount int) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	joinCmd, err := exec.Run(ctx, "kubeadm token create --print-join-command --ttl=1h")
	if err != nil {
		return fmt.Errorf("failed to generate join token: %w", err)
	}
	joinCmd = strings.TrimSpace(joinCmd)

	return p.provisionScaleUpNodes(ctx, cluster, currentCount, joinCmd)
}

func (p *KubeadmProvisioner) provisionScaleUpNodes(ctx context.Context, cluster *domain.Cluster, currentCount int, joinCmd string) error {
	needed := cluster.WorkerCount - currentCount
	var errs []error
	for i := 0; i < needed; i++ {
		workerName := fmt.Sprintf("worker-scale-%d", time.Now().UnixNano())
		if err := p.provisionSingleScaleNode(ctx, cluster, workerName, joinCmd); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("scale up encountered errors: %v", errs)
	}
	return nil
}

func (p *KubeadmProvisioner) provisionSingleScaleNode(ctx context.Context, cluster *domain.Cluster, workerName, joinCmd string) error {
	worker, err := p.createNode(ctx, cluster, workerName, domain.NodeRoleWorker)
	if err != nil {
		return fmt.Errorf("failed to create node %s: %w", workerName, err)
	}

	ip := p.waitForIP(ctx, worker.ID)
	if ip == "" {
		if delErr := p.repo.DeleteNode(ctx, worker.ID); delErr != nil {
			p.logger.Error("failed to cleanup orphaned node", "node_id", worker.ID, "error", delErr)
		}
		return fmt.Errorf("node %s failed to get IP", workerName)
	}

	if err := p.bootstrapNode(ctx, cluster, ip, cluster.Version, false); err != nil {
		if termErr := p.instSvc.TerminateInstance(ctx, worker.ID.String()); termErr != nil {
			p.logger.Error("failed to terminate instance after bootstrap error", "instance_id", worker.ID, "error", termErr)
		}
		if delErr := p.repo.DeleteNode(ctx, worker.ID); delErr != nil {
			p.logger.Error("failed to delete node record after bootstrap error", "node_id", worker.ID, "error", delErr)
		}
		return fmt.Errorf("failed to bootstrap node %s (%s): %w", workerName, ip, err)
	}

	if err := p.joinCluster(ctx, cluster, ip, joinCmd); err != nil {
		if err := p.updateNodeStatus(ctx, cluster.ID, worker.ID, "failed"); err != nil {
			p.logger.Error("failed to update node status to failed", "node_id", worker.ID, "error", err)
		}
		return fmt.Errorf("failed to join node %s to cluster: %w", workerName, err)
	}

	if err := p.updateNodeStatus(ctx, cluster.ID, worker.ID, "active"); err != nil {
		p.logger.Error("failed to update node status to active", "node_id", worker.ID, "error", err)
		return fmt.Errorf("node %s active but status update failed: %w", workerName, err)
	}
	return nil
}

func (p *KubeadmProvisioner) scaleDown(ctx context.Context, cluster *domain.Cluster, workers []*domain.ClusterNode) error {
	toDelete := len(workers) - cluster.WorkerCount
	var errs []error

	for i := 0; i < toDelete; i++ {
		node := workers[i]
		// In a real system, we should drain the node first
		p.logger.Info("scaling down: deleting worker", "instance_id", node.InstanceID)

		if err := p.instSvc.TerminateInstance(ctx, node.InstanceID.String()); err != nil {
			p.logger.Error("failed to terminate instance", "instance_id", node.InstanceID, "error", err)
			errs = append(errs, fmt.Errorf("failed to terminate instance %s: %w", node.InstanceID, err))
		}

		if err := p.repo.DeleteNode(ctx, node.ID); err != nil {
			p.logger.Error("failed to delete node record", "node_id", node.ID, "error", err)
			errs = append(errs, fmt.Errorf("failed to delete node %s: %w", node.ID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("scale down encountered errors: %v", errs)
	}
	return nil
}

func (p *KubeadmProvisioner) GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error) {
	if len(cluster.ControlPlaneIPs) == 0 {
		return "", fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return "", err
	}

	if role == "viewer" {
		// Create a limited kubeconfig for viewers
		// Implementation deferred for v2.2 plan
		return "", fmt.Errorf("viewer role not yet implemented")
	}

	return exec.Run(ctx, "cat "+adminKubeconfig)
}

func (p *KubeadmProvisioner) GetHealth(ctx context.Context, cluster *domain.Cluster) (*ports.ClusterHealth, error) {
	if len(cluster.ControlPlaneIPs) == 0 {
		return nil, fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return nil, err
	}

	health := &ports.ClusterHealth{
		Status: cluster.Status,
	}

	// 1. Check API Server
	_, err = exec.Run(ctx, kubectlBase+" get nodes")
	health.APIServer = (err == nil)

	// 2. Node Counts
	nodesOut, err := exec.Run(ctx, kubectlBase+" get nodes --no-headers")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(nodesOut), "\n")
		health.NodesTotal = len(lines)
		readyCount := 0
		for _, line := range lines {
			if strings.Contains(line, " Ready ") {
				readyCount++
			}
		}
		health.NodesReady = readyCount
	}

	if !health.APIServer {
		health.Message = "API server is unreachable"
	} else if health.NodesReady < health.NodesTotal {
		health.Message = fmt.Sprintf("%d/%d nodes are ready", health.NodesReady, health.NodesTotal)
	}

	return health, nil
}
