package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) initKubeadm(ctx context.Context, cluster *domain.Cluster, ip, publicIP string) (joinCmd string, kubeconfig string, err error) {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return "", "", err
	}

	// Init command
	initCmd := fmt.Sprintf("kubeadm init --kubernetes-version=%s --pod-network-cidr=%s --apiserver-cert-extra-sans=%s --ignore-preflight-errors=all", cluster.Version, podCIDR, publicIP)
	out, err := exec.Run(ctx, initCmd)
	if err != nil {
		return "", "", err
	}

	// Extract join command (more robust multi-line capture)
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if strings.Contains(line, "kubeadm join") {
			cmd := line
			for j := i + 1; j < len(lines); j++ {
				trimmed := strings.TrimSpace(lines[j])
				if trimmed == "" {
					break
				}
				cmd += " " + trimmed
			}
			joinCmd = strings.TrimSpace(strings.ReplaceAll(cmd, "\\", ""))
			p.logger.Info("extracted join command", "command", joinCmd)
			break
		}
	}

	// Extract kubeconfig
	kubeconfig, err = exec.Run(ctx, "cat "+adminKubeconfig)
	if err != nil {
		return "", "", err
	}

	return joinCmd, kubeconfig, nil
}

func (p *KubeadmProvisioner) joinCluster(ctx context.Context, cluster *domain.Cluster, ip, joinCmd string) error {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return err
	}
	var joinErr error
	for i := 0; i < 3; i++ {
		_, joinErr = exec.Run(ctx, joinCmd+" --ignore-preflight-errors=all")
		if joinErr == nil {
			return nil
		}
		p.logger.Warn("failed to join cluster, retrying...", "ip", ip, "error", joinErr, "attempt", i+1)
		time.Sleep(10 * time.Second)
	}
	return joinErr
}
