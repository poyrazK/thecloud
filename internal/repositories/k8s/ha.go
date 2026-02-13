package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (p *KubeadmProvisioner) ensureAPIServerLB(ctx context.Context, cluster *domain.Cluster) error {
	lbName := fmt.Sprintf("lb-k8s-%s", cluster.Name)

	// Check if already exists
	existingLBs, err := p.lbSvc.List(ctx)
	if err == nil {
		for _, lb := range existingLBs {
			if lb.Name == lbName && lb.VpcID == cluster.VpcID {
				cluster.APIServerLBAddress = &lb.IP
				return nil
			}
		}
	}

	p.logger.Info("creating API server load balancer", "lb_name", lbName)
	lb, err := p.lbSvc.Create(ctx, lbName, cluster.VpcID, 6443, "round-robin", cluster.ID.String())
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %w", err)
	}

	// Wait for IP
	lbIP := ""
	for i := 0; i < 10; i++ {
		if lb.IP != "" {
			lbIP = lb.IP
			break
		}
		time.Sleep(2 * time.Second)
		if lbNew, err := p.lbSvc.Get(ctx, lb.ID.String()); err == nil {
			lb = lbNew
		}
	}

	if lbIP == "" {
		return fmt.Errorf("load balancer IP never assigned")
	}

	cluster.APIServerLBAddress = &lbIP
	return p.repo.Update(ctx, cluster)
}
