package ccm

import (
	"context"
	"fmt"
	"strings"

	"github.com/poyrazk/thecloud/pkg/sdk"
	v1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

const (
	DefaultLBAlgorithm  = "round-robin"
	DefaultTargetWeight = 1
)

type loadbalancer struct {
	client *sdk.Client
}

func newLoadBalancer(client *sdk.Client) cloudprovider.LoadBalancer {
	return &loadbalancer{
		client: client,
	}
}

func (l *loadbalancer) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	// LB names must be unique within the tenant, so include cluster and service details.
	return fmt.Sprintf("k8s-%s-%s-%s", clusterName, service.Namespace, service.Name)
}

func (l *loadbalancer) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	name := l.GetLoadBalancerName(ctx, clusterName, service)

	lb, err := l.findLBByName(ctx, name)
	if err != nil {
		return nil, false, err
	}
	if lb == nil {
		return nil, false, nil
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				// In a real scenario we'd return the LB's external IP or DNS name.
				// Currently, The Cloud API returns a LoadBalancer struct.
				// For the sake of the CCM, if there's no public IP field yet, we fallback to its name or a dummy.
				Hostname: fmt.Sprintf("%s.thecloud.local", lb.ID),
			},
		},
	}, true, nil
}

func (l *loadbalancer) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	name := l.GetLoadBalancerName(ctx, clusterName, service)
	klog.Infof("EnsureLoadBalancer(%s)", name)

	// Validate service
	if len(service.Spec.Ports) == 0 {
		return nil, fmt.Errorf("service has no ports")
	}

	// We map the first port as a simplification.
	// A production CCM would handle multiple ports by creating multiple LBs or listeners.
	svcPort := service.Spec.Ports[0]
	port := int(svcPort.Port)

	// Validate NodePort
	if svcPort.NodePort == 0 {
		klog.Infof("NodePort not yet allocated for service %s/%s, skipping reconciliation", service.Namespace, service.Name)
		return nil, nil
	}

	// We need the VPC ID for the LB. In our architecture, the nodes are in a VPC.
	// We can fetch one of the nodes to find its VPC ID.
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available to attach to load balancer")
	}

	inst, err := l.client.GetInstanceWithContext(ctx, nodes[0].Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance for node %s to determine VPC: %w", nodes[0].Name, err)
	}
	vpcID := inst.VpcID
	if vpcID == "" {
		return nil, fmt.Errorf("node %s is not in a VPC", nodes[0].Name)
	}

	lb, err := l.findLBByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if lb == nil {
		klog.Infof("Creating new LB %s in VPC %s for port %d", name, vpcID, port)
		lb, err = l.client.CreateLBWithContext(ctx, name, vpcID, port, DefaultLBAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("failed to create LB: %w", err)
		}
	}

	// Reconcile targets
	err = l.reconcileTargets(ctx, lb.ID, svcPort.NodePort, nodes)
	if err != nil {
		return nil, err
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				Hostname: fmt.Sprintf("%s.thecloud.local", lb.ID),
			},
		},
	}, nil
}

func (l *loadbalancer) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	name := l.GetLoadBalancerName(ctx, clusterName, service)
	klog.Infof("UpdateLoadBalancer(%s)", name)

	lb, err := l.findLBByName(ctx, name)
	if err != nil {
		return err
	}
	if lb == nil {
		return fmt.Errorf("load balancer %s not found for update", name)
	}

	if len(service.Spec.Ports) == 0 {
		return nil
	}
	svcPort := service.Spec.Ports[0]

	// Validate NodePort
	if svcPort.NodePort == 0 {
		klog.Infof("NodePort not yet allocated for service %s/%s during update, skipping reconciliation", service.Namespace, service.Name)
		return nil
	}

	return l.reconcileTargets(ctx, lb.ID, svcPort.NodePort, nodes)
}

func (l *loadbalancer) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	name := l.GetLoadBalancerName(ctx, clusterName, service)
	klog.Infof("EnsureLoadBalancerDeleted(%s)", name)

	lb, err := l.findLBByName(ctx, name)
	if err != nil {
		return err
	}
	if lb == nil {
		klog.Infof("Load balancer %s already deleted", name)
		return nil
	}

	err = l.client.DeleteLBWithContext(ctx, lb.ID)
	if err != nil {
		return fmt.Errorf("failed to delete LB %s: %w", lb.ID, err)
	}

	return nil
}

// Helpers

func (l *loadbalancer) findLBByName(ctx context.Context, name string) (*sdk.LoadBalancer, error) {
	lbs, err := l.client.ListLBsWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list LBs: %w", err)
	}

	for _, lb := range lbs {
		if lb.Name == name {
			// Found it
			// create a copy to return pointer safely
			found := lb
			return &found, nil
		}
	}
	return nil, nil
}

func (l *loadbalancer) reconcileTargets(ctx context.Context, lbID string, nodePort int32, nodes []*v1.Node) error {
	existingTargets, err := l.client.ListLBTargetsWithContext(ctx, lbID)
	if err != nil {
		return fmt.Errorf("failed to list LB targets: %w", err)
	}

	// Build sets for quick lookup
	desiredNodes := make(map[string]*v1.Node)
	for _, n := range nodes {
		desiredNodes[n.Name] = n
	}

	existingNodes := make(map[string]sdk.LBTarget)
	for _, t := range existingTargets {
		// Map instance ID back to Name is complex unless we list instances.
		// Let's list instances to correlate ID <-> Name
		existingNodes[t.InstanceID] = t
	}

	// We need instances to match Names to IDs
	instances, err := l.client.ListInstancesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances for target reconciliation: %w", err)
	}

	instanceNameToID := make(map[string]string)
	for _, inst := range instances {
		instanceNameToID[inst.Name] = inst.ID
	}

	// Add missing targets
	for nodeName := range desiredNodes {
		instID, ok := instanceNameToID[nodeName]
		if !ok {
			klog.Warningf("Node %s not found in instances list, skipping target addition", nodeName)
			continue
		}

		if _, exists := existingNodes[instID]; !exists {
			klog.Infof("Adding node %s (ID: %s) to LB %s", nodeName, instID, lbID)
			err := l.client.AddLBTargetWithContext(ctx, lbID, instID, int(nodePort), DefaultTargetWeight)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to add target %s: %w", instID, err)
			}
		}
	}

	// Remove stale targets
	desiredInstanceIDs := make(map[string]bool)
	for nodeName := range desiredNodes {
		if id, ok := instanceNameToID[nodeName]; ok {
			desiredInstanceIDs[id] = true
		}
	}

	for instID := range existingNodes {
		if !desiredInstanceIDs[instID] {
			klog.Infof("Removing stale node (ID: %s) from LB %s", instID, lbID)
			err := l.client.RemoveLBTargetWithContext(ctx, lbID, instID)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("failed to remove target %s: %w", instID, err)
			}
		}
	}

	return nil
}
