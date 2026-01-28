// Package k8s implements Kubernetes provisioning adapters.
package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) installCNI(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}
	// Install Calico CNI with retries
	calicoURL := fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/calico.yaml", calicoVersion)
	var applyErr error
	for i := 0; i < 3; i++ {
		_, applyErr = exec.Run(ctx, fmt.Sprintf(kubectlApply, calicoURL))
		if applyErr == nil {
			return nil
		}
		p.logger.Warn("failed to apply CNI, retrying...", "error", applyErr, "attempt", i+1)
		time.Sleep(15 * time.Second)
	}
	return applyErr
}

func (p *KubeadmProvisioner) patchKubeProxy(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	// In KIND/Docker environments, kube-proxy conntrack usually fails with permission denied.
	// We patch the ConfigMap to set maxPerCore: 0.
	patchCmd := kubectlBase + ` -n kube-system patch configmap kube-proxy --type='json' -p='[{"op": "replace", "path": "/data/config.conf", "value": "apiVersion: kubeproxy.config.k8s.io/v1alpha1\nkind: KubeProxyConfiguration\nmode: \"\"\nconntrack:\n  maxPerCore: 0"}]'`
	_, err = exec.Run(ctx, patchCmd)
	if err != nil {
		return err
	}

	// Restart kube-proxy pods
	restartCmd := kubectlBase + " -n kube-system delete pod -l k8s-app=kube-proxy"
	_, err = exec.Run(ctx, restartCmd)
	return err
}

func (p *KubeadmProvisioner) installObservability(ctx context.Context, cluster *domain.Cluster, masterIP string) error {
	exec, err := p.getExecutor(ctx, cluster, masterIP)
	if err != nil {
		return err
	}

	p.logger.Info("installing observability components", "ip", masterIP)

	// Install kube-state-metrics
	ksmManifest := "https://github.com/kubernetes/kube-state-metrics/releases/download/v2.10.1/standard.yaml"
	_, err = exec.Run(ctx, fmt.Sprintf(kubectlApply, ksmManifest))
	if err != nil {
		return fmt.Errorf("failed to deploy kube-state-metrics: %w", err)
	}

	return nil
}
