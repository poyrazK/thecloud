// Package k8s implements Kubernetes provisioning adapters.
package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

const (
	defaultUser     = "ubuntu"
	adminKubeconfig = "/etc/kubernetes/admin.conf"
)

// KubeadmProvisioner implements ports.ClusterProvisioner using kubeadm and Cloud-Init.
type KubeadmProvisioner struct {
	instSvc     ports.InstanceService
	repo        ports.ClusterRepository
	secretSvc   ports.SecretService
	sgSvc       ports.SecurityGroupService
	storageSvc  ports.StorageService
	lbSvc       ports.LBService
	logger      *slog.Logger
	templateDir string
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
	templateDir := "internal/repositories/k8s/templates"
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// Fallback for running tests from within the k8s repo
		if _, err := os.Stat("templates"); err == nil {
			templateDir = "templates"
		}
	}

	return &KubeadmProvisioner{
		instSvc:     instSvc,
		repo:        repo,
		secretSvc:   secretSvc,
		sgSvc:       sgSvc,
		storageSvc:  storageSvc,
		lbSvc:       lbSvc,
		logger:      logger,
		templateDir: templateDir,
	}
}

func (p *KubeadmProvisioner) Provision(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("starting real provisioning for cluster", "cluster_id", cluster.ID, "name", cluster.Name)

	// Phase 1: Ensure Security Group
	if err := p.ensureClusterSecurityGroup(ctx, cluster); err != nil {
		return p.failCluster(ctx, cluster, "failed to ensure security group", err)
	}

	// Phase 2: Handle HA LB if enabled
	if cluster.HAEnabled {
		if err := p.ensureAPIServerLB(ctx, cluster); err != nil {
			return p.failCluster(ctx, cluster, "failed to ensure API Server LB", err)
		}
	}

	// Phase 3: Provision Control Plane via Cloud-Init
	if err := p.provisionControlPlane(ctx, cluster); err != nil {
		return err
	}

	// Phase 4: Generate Join Token (requires SSH to CP)
	if err := p.refreshJoinToken(ctx, cluster); err != nil {
		return err
	}

	// Phase 5: Provision Workers via Cloud-Init (using join token)
	if err := p.provisionWorkers(ctx, cluster); err != nil {
		return err
	}

	cluster.Status = domain.ClusterStatusRunning
	_ = p.repo.Update(ctx, cluster)
	return nil
}

func (p *KubeadmProvisioner) provisionControlPlane(ctx context.Context, cluster *domain.Cluster) error {
	userData, err := p.renderTemplate("control_plane.yaml", map[string]interface{}{
		"PodCIDR":     cluster.PodCIDR,
		"ServiceCIDR": cluster.ServiceCIDR,
		"HAEnabled":   cluster.HAEnabled,
		"LBAddress":   cluster.APIServerLBAddress,
	})
	if err != nil {
		return p.failCluster(ctx, cluster, "failed to render control plane template", err)
	}

	nodeName := fmt.Sprintf("%s-cp-0", cluster.Name)
	inst, err := p.instSvc.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
		Name:      nodeName,
		ImageName: "ubuntu:22.04", // Use canonical docker hub image
		NetworkID: cluster.VpcID.String(),
		UserData:  userData,
	})
	if err != nil {
		return p.failCluster(ctx, cluster, "failed to launch control plane instance", err)
	}

	node := &domain.ClusterNode{
		ID:         uuid.New(),
		ClusterID:  cluster.ID,
		InstanceID: inst.ID,
		Role:       domain.NodeRoleControlPlane,
		Status:     "provisioning",
		JoinedAt:   time.Now(),
	}
	if err := p.repo.AddNode(ctx, node); err != nil {
		return err
	}

	// Wait for instance to have an IP
	masterIP := p.waitForIP(ctx, node.InstanceID)
	if masterIP == "" {
		return p.failCluster(ctx, cluster, "control plane node failed to get an IP", nil)
	}
	cluster.ControlPlaneIPs = append(cluster.ControlPlaneIPs, masterIP)
	_ = p.repo.Update(ctx, cluster)

	// Wait for kubeadm init to finish and kubeconfig to be available via SSH
	p.logger.Info("waiting for kubeadm init to complete", "ip", masterIP)

	// Implementation of waitForKubeconfig using SSH with retries
	kubeconfig, err := p.waitForKubeconfig(ctx, cluster, masterIP)
	if err != nil {
		return p.failCluster(ctx, cluster, "failed to retrieve kubeconfig", err)
	}

	encryptedKubeconfig, err := p.secretSvc.Encrypt(ctx, cluster.UserID, kubeconfig)
	if err != nil {
		return p.failCluster(ctx, cluster, "failed to encrypt kubeconfig", err)
	}
	cluster.KubeconfigEncrypted = encryptedKubeconfig

	_ = p.repo.Update(ctx, cluster)
	return nil
}

func (p *KubeadmProvisioner) provisionWorkers(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("provisioning worker nodes", "count", cluster.WorkerCount)

	apiServer := cluster.ControlPlaneIPs[0]
	if cluster.HAEnabled && cluster.APIServerLBAddress != nil {
		apiServer = *cluster.APIServerLBAddress
	}

	userData, err := p.renderTemplate("worker.yaml", map[string]interface{}{
		"APIServerAddress": apiServer,
		"JoinToken":        cluster.JoinToken,
		"CACertHash":       cluster.CACertHash,
	})
	if err != nil {
		return p.failCluster(ctx, cluster, "failed to render worker template", err)
	}

	for i := 0; i < cluster.WorkerCount; i++ {
		workerName := fmt.Sprintf("%s-worker-%d", cluster.Name, i)
		workerInst, err := p.instSvc.LaunchInstanceWithOptions(ctx, ports.CreateInstanceOptions{
			Name:      workerName,
			ImageName: "ubuntu-22.04",
			NetworkID: cluster.VpcID.String(),
			UserData:  userData,
		})
		if err != nil {
			p.logger.Error("failed to create worker node", "name", workerName, "error", err)
			continue
		}

		node := &domain.ClusterNode{
			ID:         uuid.New(),
			ClusterID:  cluster.ID,
			InstanceID: workerInst.ID,
			Role:       domain.NodeRoleWorker,
			Status:     "provisioning",
			JoinedAt:   time.Now(),
		}
		_ = p.repo.AddNode(ctx, node)
	}

	return nil
}

func (p *KubeadmProvisioner) renderTemplate(name string, data interface{}) (string, error) {
	path := filepath.Join(p.templateDir, name)
	tpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *KubeadmProvisioner) refreshJoinToken(ctx context.Context, cluster *domain.Cluster) error {
	if cluster.TokenExpiresAt != nil && cluster.TokenExpiresAt.After(time.Now().Add(10*time.Minute)) {
		return nil
	}

	p.logger.Info("refreshing join token", "cluster_id", cluster.ID)

	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	// Create new token
	out, err := exec.Run(ctx, "kubeadm token create --print-join-command")
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create join token", err)
	}

	// Parse join command
	// Example: kubeadm join 10.0.0.10:6443 --token abcdef.0123456789abcdef --discovery-token-ca-cert-hash sha256:hash
	fields := strings.Fields(out)
	for i, f := range fields {
		if f == "--token" && i+1 < len(fields) {
			cluster.JoinToken = fields[i+1]
		}
		if f == "--discovery-token-ca-cert-hash" && i+1 < len(fields) {
			cluster.CACertHash = fields[i+1]
		}
	}

	expiry := time.Now().Add(23 * time.Hour)
	cluster.TokenExpiresAt = &expiry

	return p.repo.Update(ctx, cluster)
}

func (p *KubeadmProvisioner) waitForIP(ctx context.Context, instID uuid.UUID) string {
	for i := 0; i < 30; i++ {
		inst, err := p.instSvc.GetInstance(ctx, instID.String())
		if err == nil && inst.PrivateIP != "" {
			ip := inst.PrivateIP
			if idx := strings.Index(ip, "/"); idx != -1 {
				ip = ip[:idx]
			}
			return ip
		}
		time.Sleep(5 * time.Second)
	}
	return ""
}

func (p *KubeadmProvisioner) waitForKubeconfig(ctx context.Context, cluster *domain.Cluster, ip string) (string, error) {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return "", err
	}

	for i := 0; i < 60; i++ { // Wait up to 10 mins
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		out, err := exec.Run(ctx, "cat "+adminKubeconfig)
		if err == nil {
			return out, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Second):
			continue
		}
	}
	return "", fmt.Errorf("timed out waiting for kubeconfig")
}

func (p *KubeadmProvisioner) getExecutor(ctx context.Context, cluster *domain.Cluster, ip string) (NodeExecutor, error) {
	// 1. Try to find instance by IP to use ServiceExecutor (preferred for Docker/Local/Managed)
	// This avoids needing SSH access if we have direct control via the backend.
	instances, err := p.instSvc.ListInstances(ctx)
	if err == nil {
		for _, inst := range instances {
			if inst.PrivateIP == ip {
				// We found the managed instance! Use the ServiceExecutor.
				return NewServiceExecutor(p.instSvc, inst.ID), nil
			}
		}
	}

	// 2. Fallback to SSH
	if cluster.SSHPrivateKeyEncrypted == "" {
		return nil, errors.New(errors.Internal, "no SSH key found for node access")
	}

	decryptedKey, err := p.secretSvc.Decrypt(ctx, cluster.UserID, cluster.SSHPrivateKeyEncrypted)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to decrypt SSH key", err)
	}

	return NewSSHExecutor(ip, defaultUser, decryptedKey), nil
}

func (p *KubeadmProvisioner) Deprovision(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("deprovisioning cluster", "cluster_id", cluster.ID)

	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		_ = p.instSvc.TerminateInstance(ctx, node.InstanceID.String())
		_ = p.repo.DeleteNode(ctx, node.ID)
	}

	if cluster.HAEnabled {
		lbName := fmt.Sprintf("lb-k8s-%s", cluster.Name)
		lbs, err := p.lbSvc.List(ctx)
		if err == nil {
			for _, lb := range lbs {
				if lb.Name == lbName && lb.VpcID == cluster.VpcID {
					p.logger.Info("cleaning up cluster load balancer", "cluster_id", cluster.ID, "lb_id", lb.ID)
					_ = p.lbSvc.Delete(ctx, lb.ID.String())
				}
			}
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

func (p *KubeadmProvisioner) GetStatus(ctx context.Context, cluster *domain.Cluster) (domain.ClusterStatus, error) {
	return cluster.Status, nil
}

func (p *KubeadmProvisioner) Repair(ctx context.Context, cluster *domain.Cluster) error { return nil }

func (p *KubeadmProvisioner) Scale(ctx context.Context, cluster *domain.Cluster) error {
	return p.provisionWorkers(ctx, cluster)
}

func (p *KubeadmProvisioner) GetKubeconfig(ctx context.Context, cluster *domain.Cluster, role string) (string, error) {
	if cluster.KubeconfigEncrypted == "" {
		return "", errors.New(errors.NotFound, "kubeconfig not found")
	}
	return p.secretSvc.Decrypt(ctx, cluster.UserID, cluster.KubeconfigEncrypted)
}

func (p *KubeadmProvisioner) GetHealth(ctx context.Context, cluster *domain.Cluster) (*ports.ClusterHealth, error) {
	if len(cluster.ControlPlaneIPs) == 0 {
		return nil, errors.New(errors.NotFound, "no control plane IPs found")
	}

	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return nil, err
	}

	// Check API Server
	_, err = exec.Run(ctx, "kubectl --kubeconfig "+adminKubeconfig+" get nodes")
	health := &ports.ClusterHealth{
		APIServer: err == nil,
	}

	if err != nil {
		health.Message = "API server is unreachable"
		//nolint:nilerr
		return health, nil
	}

	// Get Node Status
	out, err := exec.Run(ctx, "kubectl --kubeconfig "+adminKubeconfig+" get nodes --no-headers")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		health.NodesTotal = len(lines)
		for _, line := range lines {
			if strings.Contains(line, " Ready ") {
				health.NodesReady++
			}
		}
		health.Message = fmt.Sprintf("%d/%d nodes are ready", health.NodesReady, health.NodesTotal)
	}

	return health, nil
}

func (p *KubeadmProvisioner) Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error {
	p.logger.Info("upgrading cluster", "cluster_id", cluster.ID, "to_version", version)

	if len(cluster.ControlPlaneIPs) == 0 {
		return errors.New(errors.InvalidInput, "no control plane nodes found")
	}

	// 1. Upgrade Control Plane nodes one by one
	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Role == domain.NodeRoleControlPlane {
			if err := p.upgradeNode(ctx, cluster, node, version); err != nil {
				return err
			}
		}
	}

	// 2. Upgrade Worker nodes one by one
	for _, node := range nodes {
		if node.Role == domain.NodeRoleWorker {
			if err := p.upgradeNode(ctx, cluster, node, version); err != nil {
				return err
			}
		}
	}

	cluster.Version = version
	return p.repo.Update(ctx, cluster)
}

func (p *KubeadmProvisioner) upgradeNode(ctx context.Context, cluster *domain.Cluster, node *domain.ClusterNode, version string) error {
	inst, err := p.instSvc.GetInstance(ctx, node.InstanceID.String())
	if err != nil {
		return err
	}

	exec, err := p.getExecutor(ctx, cluster, inst.PrivateIP)
	if err != nil {
		return err
	}

	v := strings.TrimPrefix(version, "v")
	minor := v[:strings.LastIndex(v, ".")]

	p.logger.Info("upgrading node", "node_id", node.ID, "ip", inst.PrivateIP)

	cmds := []string{
		"apt-get update",
		fmt.Sprintf("apt-get install -y --allow-change-held-packages kubeadm=%s-*", v),
	}

	if node.Role == domain.NodeRoleControlPlane {
		cmds = append(cmds, fmt.Sprintf("kubeadm upgrade apply v%s -y", minor))
	} else {
		cmds = append(cmds, "kubeadm upgrade node")
	}

	cmds = append(cmds, fmt.Sprintf("apt-get install -y --allow-change-held-packages kubelet=%s-* kubectl=%s-*", v, v))
	cmds = append(cmds, "systemctl daemon-reload", "systemctl restart kubelet")

	for _, cmd := range cmds {
		if _, err := exec.Run(ctx, cmd); err != nil {
			return errors.Wrap(errors.Internal, "failed to run upgrade command", err)
		}
	}

	return nil
}

func (p *KubeadmProvisioner) RotateSecrets(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("rotating cluster secrets", "cluster_id", cluster.ID)

	if len(cluster.ControlPlaneIPs) == 0 {
		return errors.New(errors.InvalidInput, "no control plane node for rotation")
	}

	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	// 1. Renew certs
	if _, err := exec.Run(ctx, "kubeadm certs renew all"); err != nil {
		return err
	}

	// 2. Update Kubeconfig in DB
	kubeconfig, err := exec.Run(ctx, "cat "+adminKubeconfig)
	if err != nil {
		return err
	}

	encryptedKubeconfig, err := p.secretSvc.Encrypt(ctx, cluster.UserID, kubeconfig)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to encrypt kubeconfig", err)
	}
	cluster.KubeconfigEncrypted = encryptedKubeconfig

	return p.repo.Update(ctx, cluster)
}

func (p *KubeadmProvisioner) CreateBackup(ctx context.Context, cluster *domain.Cluster) error {
	p.logger.Info("creating cluster backup", "cluster_id", cluster.ID)

	if len(cluster.ControlPlaneIPs) == 0 {
		return errors.New(errors.InvalidInput, "no control plane node for backup")
	}

	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	// Backup etcd
	backupCmd := "ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 " +
		"--cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt " +
		"--key=/etc/kubernetes/pki/etcd/server.key snapshot save /tmp/snapshot.db"

	if _, err := exec.Run(ctx, backupCmd); err != nil {
		return errors.Wrap(errors.Internal, "etcd snapshot failed", err)
	}

	// Upload to storage service
	// We'd need a way to stream files or cat them
	snapshotData, err := exec.Run(ctx, "base64 /tmp/snapshot.db")
	if err != nil {
		return err
	}

	if p.storageSvc != nil {
		key := fmt.Sprintf("k8s-backups/%s/%d.db.b64", cluster.ID, time.Now().Unix())
		_, err = p.storageSvc.Upload(ctx, "k8s-backups", key, strings.NewReader(snapshotData))
		return err
	}

	return nil
}

func (p *KubeadmProvisioner) Restore(ctx context.Context, cluster *domain.Cluster, path string) error {
	p.logger.Info("restoring cluster from backup", "cluster_id", cluster.ID, "path", path)
	// This is a complex operation involving stopping etcd, restoring from snapshot, and restarting.
	// Implementing a full restore requires careful node-by-node handling.
	return errors.New(errors.NotImplemented, "restore not yet fully implemented")
}
