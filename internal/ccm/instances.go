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
	StatusStopped    = "STOPPED"
	StatusTerminated = "TERMINATED"
	MachineTypeStd1  = "standard-1"
)

type instances struct {
	client *sdk.Client
}

func newInstancesV2(client *sdk.Client) cloudprovider.InstancesV2 {
	return &instances{
		client: client,
	}
}

// InstanceExists returns true if the instance for the given node exists.
func (i *instances) InstanceExists(ctx context.Context, node *v1.Node) (bool, error) {
	inst, err := i.getInstance(ctx, node)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return inst != nil, nil
}

// InstanceShutdown returns true if the instance is shutdown according to the cloud provider.
func (i *instances) InstanceShutdown(ctx context.Context, node *v1.Node) (bool, error) {
	inst, err := i.getInstance(ctx, node)
	if err != nil {
		return false, err
	}
	// Assuming Status == "STOPPED" or "TERMINATED" means shutdown
	return inst.Status == StatusStopped || inst.Status == StatusTerminated, nil
}

// InstanceMetadata returns the instance's metadata.
func (i *instances) InstanceMetadata(ctx context.Context, node *v1.Node) (*cloudprovider.InstanceMetadata, error) {
	inst, err := i.getInstance(ctx, node)
	if err != nil {
		return nil, err
	}

	addresses := []v1.NodeAddress{}

	// Private IP (Internal)
	if inst.PrivateIP != "" {
		addresses = append(addresses, v1.NodeAddress{
			Type:    v1.NodeInternalIP,
			Address: inst.PrivateIP,
		})
	}

	// Add Hostname
	addresses = append(addresses, v1.NodeAddress{
		Type:    v1.NodeHostName,
		Address: inst.Name,
	})

	// InstanceType
	instanceType := inst.InstanceType
	if instanceType == "" {
		instanceType = MachineTypeStd1
	}

	return &cloudprovider.InstanceMetadata{
		ProviderID:    fmt.Sprintf("%s://%s", ProviderName, inst.ID),
		InstanceType:  instanceType,
		NodeAddresses: addresses,
		Zone:          "local", // TheCloud currently doesn't have multiple zones, default to local
		Region:        "local",
	}, nil
}

// Helper to get an instance by node's ProviderID or Name
func (i *instances) getInstance(ctx context.Context, node *v1.Node) (*sdk.Instance, error) {
	id := ""
	if node.Spec.ProviderID != "" {
		id = strings.TrimPrefix(node.Spec.ProviderID, ProviderName+"://")
	} else {
		id = node.Name
	}

	klog.V(4).Infof("Resolving instance for node %s (resolved ID/Name: %s)", node.Name, id)

	// Propagate context to SDK call
	inst, err := i.client.GetInstanceWithContext(ctx, id)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

func isNotFound(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404")
}
