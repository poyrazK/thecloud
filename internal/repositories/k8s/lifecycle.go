package k8s

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) Upgrade(ctx context.Context, cluster *domain.Cluster, version string) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	p.logger.Info("starting Kubernetes upgrade", "cluster_id", cluster.ID, "target_version", version)

	// 1. Upgrade Control Plane
	if err := p.upgradeControlPlane(ctx, cluster, masterIP, version); err != nil {
		return fmt.Errorf("control plane upgrade failed: %w", err)
	}

	if err := p.upgradeWorkerNodes(ctx, cluster, version); err != nil {
		return err
	}

	p.logger.Info("Kubernetes upgrade completed", "cluster_id", cluster.ID, "version", version)
	return nil
}

func (p *KubeadmProvisioner) upgradeWorkerNodes(ctx context.Context, cluster *domain.Cluster, version string) error {
	nodes, err := p.repo.GetNodes(ctx, cluster.ID)
	if err != nil {
		return err
	}

	var upgradeErrs []error
	for _, node := range nodes {
		if node.Role == domain.NodeRoleWorker {
			if err := p.upgradeSingleWorkerNode(ctx, cluster, node, version); err != nil {
				upgradeErrs = append(upgradeErrs, err)
			}
		}
	}

	if len(upgradeErrs) > 0 {
		return fmt.Errorf("upgrade completed with errors: %v", upgradeErrs)
	}

	return nil
}

func (p *KubeadmProvisioner) upgradeSingleWorkerNode(ctx context.Context, cluster *domain.Cluster, node *domain.ClusterNode, version string) error {
	inst, err := p.instSvc.GetInstance(ctx, node.InstanceID.String())
	if err != nil {
		p.logger.Error("failed to get worker instance for upgrade", "instance_id", node.InstanceID, "error", err)
		return fmt.Errorf("failed to get instance for node %s: %w", node.ID, err)
	}

	ip := inst.PrivateIP
	if idx := strings.Index(ip, "/"); idx != -1 {
		ip = ip[:idx]
	}

	p.logger.Info("upgrading worker node", "ip", ip)
	if err := p.upgradeWorkerNode(ctx, cluster, ip, version); err != nil {
		p.logger.Error("worker upgrade failed", "ip", ip, "error", err)
		return fmt.Errorf("failed to upgrade node %s (%s): %w", node.ID, ip, err)
	}
	return nil
}

func (p *KubeadmProvisioner) upgradeControlPlane(ctx context.Context, cluster *domain.Cluster, ip, version string) error {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return err
	}

	ver := strings.TrimPrefix(version, "v")

	upgradeScript := fmt.Sprintf(`
set -e
# 1. Update kubeadm
apt-mark unhold kubeadm
apt-get update && apt-get install -y kubeadm=%s-1.1
apt-mark hold kubeadm

# 2. Apply upgrade
kubeadm upgrade apply %s -y

# 3. Update kubelet and kubectl
apt-mark unhold kubelet kubectl
apt-get update && apt-get install -y kubelet=%s-1.1 kubectl=%s-1.1
apt-mark hold kubelet kubectl

# 4. Restart kubelet
systemctl daemon-reload
systemctl restart kubelet
`, ver, version, ver, ver)

	p.logger.Info("running control plane upgrade script", "ip", ip, "version", version)
	_, err = exec.Run(ctx, upgradeScript)
	return err
}

func (p *KubeadmProvisioner) upgradeWorkerNode(ctx context.Context, cluster *domain.Cluster, ip, version string) error {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return err
	}

	ver := strings.TrimPrefix(version, "v")

	upgradeScript := fmt.Sprintf(`
set -e
# 1. Update kubeadm
apt-mark unhold kubeadm
apt-get update && apt-get install -y kubeadm=%s-1.1
apt-mark hold kubeadm

# 2. Upgrade node config
kubeadm upgrade node

# 3. Update kubelet and kubectl
apt-mark unhold kubelet kubectl
apt-get update && apt-get install -y kubelet=%s-1.1 kubectl=%s-1.1
apt-mark hold kubelet kubectl

# 4. Restart kubelet
systemctl daemon-reload
systemctl restart kubelet
`, ver, ver, ver)

	p.logger.Info("running worker upgrade script", "ip", ip, "version", version)
	_, err = exec.Run(ctx, upgradeScript)
	return err
}

func (p *KubeadmProvisioner) RotateSecrets(ctx context.Context, cluster *domain.Cluster) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	p.logger.Info("rotating cluster certificates", "cluster_id", cluster.ID)

	// 1. Renew all certificates
	_, err = exec.Run(ctx, "kubeadm certs renew all")
	if err != nil {
		return fmt.Errorf("failed to renew certificates: %w", err)
	}

	// 2. Re-read the admin kubeconfig
	kubeconfig, err := exec.Run(ctx, "cat "+adminKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to read renewed kubeconfig: %w", err)
	}

	// 3. Encrypt and store the new kubeconfig
	// 3. Encrypt and store the new kubeconfig
	encryptedKubeconfig, err := p.secretSvc.Encrypt(ctx, cluster.UserID, kubeconfig)
	if err != nil {
		p.logger.Error("failed to encrypt renewed kubeconfig", "cluster_id", cluster.ID, "error", err)
		return fmt.Errorf("failed to encrypt kubeconfig: %w", err)
	}
	cluster.Kubeconfig = encryptedKubeconfig

	// 4. Update repo
	return p.repo.Update(ctx, cluster)
}

func (p *KubeadmProvisioner) CreateBackup(ctx context.Context, cluster *domain.Cluster) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	tempPath := fmt.Sprintf("/tmp/snapshot-%d.db", time.Now().Unix())
	p.logger.Info("creating etcd snapshot", "cluster_id", cluster.ID, "path", tempPath)

	backupCmd := fmt.Sprintf("ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key snapshot save %s", tempPath)

	if _, err = exec.Run(ctx, backupCmd); err != nil {
		return fmt.Errorf("failed to create etcd snapshot: %w", err)
	}

	// 2. Extract snapshot as Base64 to handle binary data through the executor
	p.logger.Info("uploading snapshot to remote storage", "cluster_id", cluster.ID)
	b64Output, err := exec.Run(ctx, fmt.Sprintf("base64 -w 0 %s", tempPath))
	if err != nil {
		return fmt.Errorf("failed to read snapshot: %w", err)
	}

	snapshotData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Output))
	if err != nil {
		return fmt.Errorf("failed to decode base64 snapshot: %w", err)
	}

	bucket := "k8s-backups"
	key := fmt.Sprintf("%s/snapshot-%d.db", cluster.ID, time.Now().Unix())
	_, err = p.storageSvc.Upload(ctx, bucket, key, bytes.NewReader(snapshotData))
	if err != nil {
		return fmt.Errorf("failed to upload snapshot to storage: %w", err)
	}

	// Cleanup temp file
	// Cleanup temp file
	if _, err := exec.Run(ctx, "rm "+tempPath); err != nil {
		p.logger.Warn("failed to remove temp backup file", "path", tempPath, "error", err)
	}

	p.logger.Info("etcd snapshot backed up to remote storage", "key", key)
	return nil
}

func (p *KubeadmProvisioner) Restore(ctx context.Context, cluster *domain.Cluster, backupPath string) error {
	if len(cluster.ControlPlaneIPs) == 0 {
		return fmt.Errorf(errNoControlPlaneIPs, cluster.ID)
	}
	masterIP := cluster.ControlPlaneIPs[0]
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	// 1. Download snapshot from StorageService if the path looks like a storage key
	// In the current system, backupPath passed from service layer is usually the storage key
	bucket := "k8s-backups"
	rc, _, err := p.storageSvc.Download(ctx, bucket, backupPath)
	if err != nil {
		return fmt.Errorf("failed to download backup from storage: %w", err)
	}
	defer rc.Close()

	snapshotBytes, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	// 2. Upload to master node via base64
	p.logger.Info("injecting snapshot into master node", "cluster_id", cluster.ID)
	remoteTemp := "/tmp/restore-snapshot.db"
	b64Data := base64.StdEncoding.EncodeToString(snapshotBytes)

	// We use a temporary file to avoid command length limits if the snapshot is large,
	// but for KIND etcd snapshots are small.
	_, err = exec.Run(ctx, fmt.Sprintf("echo '%s' | base64 -d > %s", b64Data, remoteTemp))
	if err != nil {
		return fmt.Errorf("failed to write snapshot to node: %w", err)
	}

	p.logger.Info("restoring etcd from injected snapshot", "cluster_id", cluster.ID)

	restoreScript := fmt.Sprintf(`
set -e
# 1. Stop kubelet and etcd
systemctl stop kubelet
# Move etcd manifest if it exists to ensure it stops
if [ -f /etc/kubernetes/manifests/etcd.yaml ]; then
  mv /etc/kubernetes/manifests/etcd.yaml /etc/kubernetes/etcd.yaml.bak
fi

# 2. Backup existing etcd data
mv /var/lib/etcd /var/lib/etcd-old-%[1]d || true

# 3. Restore from snapshot
ETCDCTL_API=3 etcdctl snapshot restore %[2]s \
  --data-dir /var/lib/etcd \
  --name $(hostname) \
  --initial-cluster $(hostname)=https://127.0.0.1:2380 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-advertise-peer-urls https://127.0.0.1:2380

# 4. Set permissions
chown -R root:root /var/lib/etcd

# 5. Restore manifest and start kubelet
if [ -f /etc/kubernetes/etcd.yaml.bak ]; then
  mv /etc/kubernetes/etcd.yaml.bak /etc/kubernetes/manifests/etcd.yaml
fi
systemctl start kubelet
`, time.Now().Unix(), remoteTemp)

	_, err = exec.Run(ctx, restoreScript)
	if err != nil {
		return fmt.Errorf("failed to restore etcd snapshot: %w", err)
	}

	p.logger.Info("etcd restoration completed successfully")
	return nil
}
