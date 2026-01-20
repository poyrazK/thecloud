package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	defaultUser          = "ubuntu"
	errNoControlPlaneIPs = "cluster %s has no control plane IPs"
	podCIDR              = "192.168.0.0/16" // Calico default
	// AnyCIDR represents all IPv4 addresses.
	AnyCIDR = "0.0.0.0/0"
	// adminKubeconfig is the default path for kubeconfig on control plane nodes.
	adminKubeconfig = "/etc/kubernetes/admin.conf"
	kubectlBase     = "kubectl --kubeconfig " + adminKubeconfig
	kubectlApply    = kubectlBase + " apply -f %s"
	calicoVersion   = "v3.27.0"
)

// KubeadmProvisioner implements ports.ClusterProvisioner using kubeadm and SSH.
type KubeadmProvisioner struct {
	instSvc    ports.InstanceService
	repo       ports.ClusterRepository
	secretSvc  ports.SecretService // To encrypt/decrypt the SSH key
	sgSvc      ports.SecurityGroupService
	storageSvc ports.StorageService
	lbSvc      ports.LBService
	logger     *slog.Logger
}

// NewKubeadmProvisioner constructs a new KubeadmProvisioner.
func NewKubeadmProvisioner(
	instSvc ports.InstanceService,
	repo ports.ClusterRepository,
	secretSvc ports.SecretService,
	sgSvc ports.SecurityGroupService,
	storageSvc ports.StorageService,
	lbSvc ports.LBService,
	logger *slog.Logger,
) *KubeadmProvisioner {
	return &KubeadmProvisioner{
		instSvc:    instSvc,
		repo:       repo,
		secretSvc:  secretSvc,
		sgSvc:      sgSvc,
		storageSvc: storageSvc,
		lbSvc:      lbSvc,
		logger:     logger,
	}
}

func (p *KubeadmProvisioner) GetSecretService() ports.SecretService {
	return p.secretSvc
}

func (p *KubeadmProvisioner) Provision(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("starting provisioning for cluster", "cluster_id", cluster.ID, "name", cluster.Name)

	// Phase 0: Ensure Security Group
	if err := p.ensureClusterSecurityGroup(ctx, cluster); err != nil {
		return p.failCluster(ctx, cluster, "failed to ensure security group", err)
	}

	// Phase 1: Control Plane
	joinCmd, err := p.provisionControlPlane(ctx, cluster)
	if err != nil {
		return err
	}

	// Phase 2: Workers
	_, err = p.provisionWorkers(ctx, cluster, joinCmd)
	if err != nil {
		return err
	}

	// Phase 3: Finalize (CNI, etc)
	masterIP := cluster.ControlPlaneIPs[0]
	return p.finalizeCluster(ctx, cluster, masterIP)
}

func (p *KubeadmProvisioner) provisionControlPlane(ctx context.Context, cluster *domain.Cluster) (string, error) {
	if cluster.HAEnabled {
		return p.provisionHAControlPlane(ctx, cluster)
	}

	master, err := p.createNode(ctx, cluster, "master-0", domain.NodeRoleControlPlane)
	if err != nil {
		return "", p.failCluster(ctx, cluster, "failed to create master node", err)
	}

	masterIP := p.waitForIP(ctx, master.ID)
	if masterIP == "" {
		return "", p.failCluster(ctx, cluster, "master node failed to get an IP", nil)
	}
	cluster.ControlPlaneIPs = []string{masterIP}
	_ = p.repo.Update(ctx, cluster)

	p.logger.Info("bootstrapping master node", "ip", masterIP)
	if err := p.bootstrapNode(ctx, cluster, masterIP, cluster.Version, true); err != nil {
		return "", p.failCluster(ctx, cluster, "failed to bootstrap master node", err)
	}

	joinCmd, kubeconfig, err := p.initKubeadm(ctx, cluster, masterIP, cluster.ControlPlaneIPs[0])
	if err != nil {
		return "", p.failCluster(ctx, cluster, "failed to init kubeadm", err)
	}

	// Wait for API server to be healthy before proceeding
	p.logger.Info("waiting for API server to become healthy", "ip", masterIP)
	if err := p.waitForAPIServer(ctx, cluster, masterIP); err != nil {
		return "", p.failCluster(ctx, cluster, "API server failed to become healthy", err)
	}

	encryptedKubeconfig, err := p.secretSvc.Encrypt(ctx, cluster.UserID, kubeconfig)
	if err != nil {
		p.logger.Error("failed to encrypt kubeconfig", "cluster_id", cluster.ID, "error", err)
		cluster.Kubeconfig = kubeconfig
	} else {
		cluster.Kubeconfig = encryptedKubeconfig
	}
	_ = p.repo.Update(ctx, cluster)

	return joinCmd, nil
}

func (p *KubeadmProvisioner) provisionWorkers(ctx context.Context, cluster *domain.Cluster, joinCmd string) (string, error) {
	for i := 0; i < cluster.WorkerCount; i++ {
		workerName := fmt.Sprintf("worker-%d", i)
		worker, err := p.createNode(ctx, cluster, workerName, domain.NodeRoleWorker)
		if err != nil {
			p.logger.Error("failed to create worker node", "name", workerName, "error", err)
			continue
		}

		workerIP := p.waitForIP(ctx, worker.ID)
		if workerIP == "" {
			p.logger.Error("worker node failed to get an IP", "name", workerName)
			continue
		}

		if err := p.bootstrapNode(ctx, cluster, workerIP, cluster.Version, false); err != nil {
			p.logger.Error("failed to bootstrap worker node", "ip", workerIP, "error", err)
			continue
		}

		if err := p.joinCluster(ctx, cluster, workerIP, joinCmd); err != nil {
			p.logger.Error("failed to join worker to cluster", "ip", workerIP, "error", err)
			_ = p.updateNodeStatus(ctx, cluster.ID, worker.ID, "failed")
			continue
		}

		_ = p.updateNodeStatus(ctx, cluster.ID, worker.ID, "active")
	}
	return joinCmd, nil
}

func (p *KubeadmProvisioner) updateNodeStatus(ctx context.Context, clusterID, instID uuid.UUID, status string) error {
	nodes, err := p.repo.GetNodes(ctx, clusterID)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if n.InstanceID == instID {
			n.Status = status
			return p.repo.UpdateNode(ctx, n)
		}
	}
	return nil
}

func (p *KubeadmProvisioner) finalizeCluster(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	p.logger.Info("finalizing cluster state", "ip", masterIP)

	// 1. Install Calico CNI
	if err := p.installCNI(ctx, cluster, masterIP); err != nil {
		return p.failCluster(ctx, cluster, "failed to install CNI", err)
	}

	// 2. Patch kube-proxy for Docker/conntrack
	if err := p.patchKubeProxy(ctx, cluster, masterIP); err != nil {
		p.logger.Error("failed to patch kube-proxy", "cluster_id", cluster.ID, "error", err)
	}

	// 3. Apply base security policies
	if err := p.applyBaseSecurity(ctx, cluster, masterIP); err != nil {
		p.logger.Error("failed to apply base security manifests", "cluster_id", cluster.ID, "error", err)
	}

	// 4. Install Observability (kube-state-metrics)
	if err := p.installObservability(ctx, cluster, masterIP); err != nil {
		p.logger.Error("failed to install observability components", "cluster_id", cluster.ID, "error", err)
	}

	cluster.Status = domain.ClusterStatusRunning
	_ = p.repo.Update(ctx, cluster)
	p.logger.Info("cluster finalized and running", "cluster_id", cluster.ID)
	return nil
}

func (p *KubeadmProvisioner) getExecutor(ctx context.Context, cluster *domain.Cluster, ip string) (NodeExecutor, error) {
	// Try to find instance by IP to determine if we can use ServiceExecutor (e.g., Docker backend)
	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err == nil {
		for _, node := range nodes {
			inst, err := p.instSvc.GetInstance(ctx, node.InstanceID.String())
			if err == nil {
				// Match by IP (strip CIDR suffix for comparison)
				instIP := inst.PrivateIP
				if idx := strings.Index(instIP, "/"); idx != -1 {
					instIP = instIP[:idx]
				}
				if instIP == ip {
					return NewServiceExecutor(p.instSvc, node.InstanceID), nil
				}
			}
		}
	}

	// Fallback to SSH
	return NewSSHExecutor(ip, defaultUser, cluster.SSHKey), nil
}

func (p *KubeadmProvisioner) Deprovision(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("deprovisioning cluster", "cluster_id", cluster.ID)

	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		p.logger.Info("terminating node instance", "instance_id", node.InstanceID)
		if err := p.instSvc.TerminateInstance(ctx, node.InstanceID.String()); err != nil {
			p.logger.Error("failed to terminate node instance", "instance_id", node.InstanceID, "error", err)
		}
		_ = p.repo.DeleteNode(ctx, node.ID)
	}

	// Optional: Delete Security Group if we created it exclusively for the cluster
	sgName := fmt.Sprintf("sg-%s", cluster.Name)
	sg, err := p.sgSvc.GetGroup(ctx, sgName, cluster.VpcID)
	if err == nil && sg != nil {
		if err := p.sgSvc.DeleteGroup(ctx, sg.ID); err != nil {
			p.logger.Error("failed to delete cluster security group", "sg_id", sg.ID, "error", err)
		}
	}

	return nil
}

func (p *KubeadmProvisioner) failCluster(ctx context.Context, cluster *domain.Cluster, msg string, err error) error {
	cluster.Status = domain.ClusterStatusFailed
	_ = p.repo.Update(ctx, cluster)
	p.logger.Error(msg, "cluster_id", cluster.ID, "error", err)
	return fmt.Errorf("%s: %w", msg, err)
}
