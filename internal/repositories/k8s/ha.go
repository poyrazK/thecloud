// Package k8s implements Kubernetes provisioning adapters.
package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) provisionHAControlPlane(ctx context.Context, cluster *domain.Cluster) (string, error) {
	p.logger.Info("provisioning HA control plane (3 nodes)", "cluster_id", cluster.ID)

	// 1. Create Load Balancer for the Control Plane
	lbName := fmt.Sprintf("lb-k8s-%s", cluster.Name)
	lb, err := p.lbSvc.Create(ctx, lbName, cluster.VpcID, 6443, "round-robin", cluster.ID.String())
	if err != nil {
		return "", p.failCluster(ctx, cluster, "failed to create control plane load balancer", err)
	}

	lbIP := ""
	for i := 0; i < 5; i++ {
		lbIP = lb.IP
		if lbIP != "" {
			break
		}
		time.Sleep(2 * time.Second)
		if lbNew, err := p.lbSvc.Get(ctx, lb.ID); err == nil {
			lb = lbNew
		}
	}

	if lbIP == "" {
		return "", p.failCluster(ctx, cluster, "load balancer IP never assigned", fmt.Errorf("lb %s has no IP", lb.ID))
	}

	cluster.APIServerLBAddress = &lbIP
	if err := p.repo.Update(ctx, cluster); err != nil {
		p.logger.Warn("failed to update cluster load balancer address", "cluster_id", cluster.ID, "error", err)
	}

	// 2. Provision 3 Master Nodes
	masterIPs, _, err := p.provisionHAMasters(ctx, cluster, lb)
	if err != nil {
		return "", err
	}

	cluster.ControlPlaneIPs = masterIPs
	if err := p.repo.Update(ctx, cluster); err != nil {
		p.logger.Warn("failed to update cluster control plane IPs", "cluster_id", cluster.ID, "error", err)
	}

	// 3. Init Kubeadm on master-0
	p.logger.Info("initializing kubeadm on first master", "ip", masterIPs[0])
	joinCmd, cpJoinCmd, kubeconfig, err := p.initKubeadmHA(ctx, cluster, masterIPs[0], lbIP)
	if err != nil {
		return "", p.failCluster(ctx, cluster, "failed to init primary master", err)
	}

	if joinCmd == "" || cpJoinCmd == "" {
		return "", p.failCluster(ctx, cluster, "failed to parse join commands from kubeadm init", fmt.Errorf("empty join commands"))
	}

	// 4. Join master-1 and master-2 to control plane
	if err := p.joinHAMasters(ctx, cluster, masterIPs, cpJoinCmd); err != nil {
		return "", err
	}

	// 5. Encrypt and store kubeconfig
	if err := p.storeKubeconfig(ctx, cluster, kubeconfig); err != nil {
		return "", p.failCluster(ctx, cluster, "failed to store HA kubeconfig", err)
	}

	return joinCmd, nil
}

func (p *KubeadmProvisioner) provisionHAMasters(ctx context.Context, cluster *domain.Cluster, lb *domain.LoadBalancer) ([]string, []*domain.Instance, error) {
	var masters []*domain.Instance
	var masterIPs []string

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("master-%d", i)
		node, err := p.createNode(ctx, cluster, name, domain.NodeRoleControlPlane)
		if err != nil {
			return nil, nil, p.failCluster(ctx, cluster, "failed to create master node "+name, err)
		}
		ip := p.waitForIP(ctx, node.ID)
		if ip == "" {
			err := fmt.Errorf("missing master IP for node %s", node.ID)
			return nil, nil, p.failCluster(ctx, cluster, "timeout waiting for IP for master "+name, err)
		}
		masters = append(masters, node)
		masterIPs = append(masterIPs, ip)

		if err := p.lbSvc.AddTarget(ctx, lb.ID, node.ID, 6443, 10); err != nil {
			return nil, nil, p.failCluster(ctx, cluster, fmt.Sprintf("failed to add master %s to LB %s", node.ID, lb.ID), err)
		}
	}

	for _, ip := range masterIPs {
		if err := p.bootstrapNode(ctx, cluster, ip, cluster.Version, true); err != nil {
			return nil, nil, p.failCluster(ctx, cluster, "failed to bootstrap master "+ip, err)
		}
	}

	return masterIPs, masters, nil
}

func (p *KubeadmProvisioner) joinHAMasters(ctx context.Context, cluster *domain.Cluster, masterIPs []string, cpJoinCmd string) error {
	for i := 1; i < 3; i++ {
		p.logger.Info("joining master to control plane", "ip", masterIPs[i])
		if err := p.joinControlPlane(ctx, cluster, masterIPs[i], cpJoinCmd); err != nil {
			return p.failCluster(ctx, cluster, "failed to join master "+masterIPs[i], err)
		}
	}
	return nil
}

func (p *KubeadmProvisioner) storeKubeconfig(ctx context.Context, cluster *domain.Cluster, kubeconfig string) error {
	encryptedKubeconfig, err := p.secretSvc.Encrypt(ctx, cluster.UserID, kubeconfig)
	if err != nil {
		p.logger.Error("failed to encrypt HA kubeconfig", "cluster_id", cluster.ID, "error", err)
		return fmt.Errorf("failed to encrypt kubeconfig: %w", err)
	}
	cluster.Kubeconfig = encryptedKubeconfig
	if err := p.repo.Update(ctx, cluster); err != nil {
		return fmt.Errorf("failed to store kubeconfig in repo: %w", err)
	}
	return nil
}

func (p *KubeadmProvisioner) initKubeadmHA(ctx context.Context, cluster *domain.Cluster, ip, lbIP string) (joinCmd, cpJoinCmd, kubeconfig string, err error) {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return "", "", "", err
	}

	initCmd := fmt.Sprintf("kubeadm init --kubernetes-version=%s --pod-network-cidr=%s --control-plane-endpoint=%s:6443 --upload-certs --ignore-preflight-errors=all", cluster.Version, podCIDR, lbIP)
	out, err := exec.Run(ctx, initCmd)
	if err != nil {
		return "", "", "", err
	}

	joinCmd, cpJoinCmd = p.parseJoinCommands(out)

	kubeconfig, err = exec.Run(ctx, "cat "+adminKubeconfig)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read admin kubeconfig: %w", err)
	}
	return joinCmd, cpJoinCmd, kubeconfig, nil
}

func (p *KubeadmProvisioner) parseJoinCommands(output string) (joinCmd, cpJoinCmd string) {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if !strings.Contains(line, "kubeadm join") {
			continue
		}

		cmd := line
		for j := i + 1; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				break
			}
			cmd += " " + trimmed
		}

		fullCmd := strings.TrimSpace(strings.ReplaceAll(cmd, "\\", ""))
		if strings.Contains(fullCmd, "--control-plane") {
			cpJoinCmd = fullCmd
		} else {
			joinCmd = fullCmd
		}
	}
	return joinCmd, cpJoinCmd
}

func (p *KubeadmProvisioner) joinControlPlane(ctx context.Context, cluster *domain.Cluster, ip, joinCmd string) error {
	exec, err := p.getExecutor(ctx, cluster, ip)
	if err != nil {
		return err
	}
	_, err = exec.Run(ctx, joinCmd+" --ignore-preflight-errors=all")
	return err
}
