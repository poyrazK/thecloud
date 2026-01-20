package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) createNode(ctx context.Context, cluster *domain.Cluster, nameTag string, role domain.NodeRole) (*domain.Instance, error) {
	name := fmt.Sprintf("%s-%s", cluster.Name, nameTag)
	// Using kindest/node image for better Docker-in-Docker support (includes systemd)
	imageName := "kindest/node:v1.29.0"
	// Don't expose port 22 to host; we access via internal network or Exec.
	inst, err := p.instSvc.LaunchInstance(ctx, name, imageName, "", &cluster.VpcID, nil, nil)
	if err != nil {
		return nil, err
	}

	// Attach Security Group
	sgName := fmt.Sprintf("sg-%s", cluster.Name)
	sg, err := p.sgSvc.GetGroup(ctx, sgName, cluster.VpcID)
	if err == nil && sg != nil {
		if err := p.sgSvc.AttachToInstance(ctx, inst.ID, sg.ID); err != nil {
			p.logger.Error("failed to attach security group to node", "node_id", inst.ID, "error", err)
			// Continue, as this shouldn't block provisioning if it fails (soft fail)
		}
	} else {
		p.logger.Warn("security group not found when attaching to node", "sg_name", sgName)
	}

	node := &domain.ClusterNode{
		ID:         uuid.New(),
		ClusterID:  cluster.ID,
		InstanceID: inst.ID,
		Role:       role,
		Status:     "provisioning",
	}
	if err := p.repo.AddNode(ctx, node); err != nil {
		return nil, err
	}

	return inst, nil
}

func (p *KubeadmProvisioner) waitForIP(ctx context.Context, instID uuid.UUID) string {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ""
		case <-timeout:
			return ""
		case <-ticker.C:
			inst, err := p.instSvc.GetInstance(ctx, instID.String())
			if err == nil && inst.PrivateIP != "" {
				// Strip CIDR suffix if present (e.g., "172.18.0.23/32" -> "172.18.0.23")
				ip := inst.PrivateIP
				if idx := strings.Index(ip, "/"); idx != -1 {
					ip = ip[:idx]
				}
				p.logger.Debug("node IP allocated", "instance_id", instID, "ip", ip)
				return ip
			}
		}
	}
}

func (p *KubeadmProvisioner) waitForAPIServer(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	checkCmd := kubectlBase + " get nodes"
	var lastErr error

	// Wait up to 5 minutes
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for API server health: %w", lastErr)
		case <-ticker.C:
			_, err = exec.Run(ctx, checkCmd)
			if err == nil {
				p.logger.Info("API server is healthy", "ip", masterIP)
				return nil
			}
			lastErr = err
			p.logger.Debug("API server not ready yet, retrying...", "error", err)
		}
	}
}

func (p *KubeadmProvisioner) bootstrapNode(ctx context.Context, cluster *domain.Cluster, ip, _ string, _ bool) error {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return err
	}

	p.logger.Info("running bootstrap script", "ip", ip)

	script := fmt.Sprintf(`
set -e
# 1. Wait for systemd to be fully operational
timeout=60
while ! systemctl is-system-running > /dev/null 2>&1; do
  STATUS=$(systemctl is-system-running || true)
  if [ "$STATUS" = "degraded" ] || [ "$STATUS" = "running" ]; then break; fi
  sleep 2
  timeout=$((timeout-2))
  if [ $timeout -le 0 ]; then break; fi
done

# 2. Kernel Hardening & Networking
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
kernel.panic_on_oops                = 1
kernel.panic                       = 10
EOF
sysctl --system

# 3. Ensure Containerd is running
if ! systemctl is-active --quiet containerd; then
  systemctl restart containerd
fi

# 4. Clean up any pre-existing kubernetes config from the image
if [ -d /etc/kubernetes/pki ]; then
	rm -rf /etc/kubernetes/pki
fi
rm -f /etc/kubernetes/admin.conf /etc/kubernetes/kubelet.conf /etc/kubernetes/controller-manager.conf /etc/kubernetes/scheduler.conf

# 5. Configure Kubelet for swap
mkdir -p /etc/default
if [ -f /etc/default/kubelet ]; then
	if ! grep -q "fail-swap-on=false" /etc/default/kubelet; then
		sed -i 's/KUBELET_EXTRA_ARGS=/KUBELET_EXTRA_ARGS=--fail-swap-on=false /' /etc/default/kubelet || echo 'KUBELET_EXTRA_ARGS=--fail-swap-on=false' >> /etc/default/kubelet
	fi
else
	echo 'KUBELET_EXTRA_ARGS="--fail-swap-on=false"' > /etc/default/kubelet
fi

# 6. Reload and restart
systemctl daemon-reload
systemctl restart kubelet

# 7. Wait for kubelet to be active
timeout=15
while ! systemctl is-active --quiet kubelet; do
  sleep 1
  timeout=$((timeout-1))
  if [ $timeout -le 0 ]; then break; fi
done

# 8. Pre-pull Calico images to avoid DNS/timeout issues during CNI install
crictl pull docker.io/calico/cni:%[1]s || true
crictl pull docker.io/calico/node:%[1]s || true
crictl pull docker.io/calico/kube-controllers:%[1]s || true
`, calicoVersion)
	_, err = exec.Run(ctx, script)
	return err
}
