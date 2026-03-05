package autoscaler

import (
	"context"
	"fmt"
	"strings"

	"github.com/poyrazk/thecloud/internal/autoscaler/protos"
	"github.com/poyrazk/thecloud/pkg/sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type AutoscalerServer struct {
	protos.UnimplementedCloudProviderServer
	client    *sdk.Client
	clusterID string
}

func NewAutoscalerServer(client *sdk.Client, clusterID string) *AutoscalerServer {
	return &AutoscalerServer{
		client:    client,
		clusterID: clusterID,
	}
}

func (s *AutoscalerServer) NodeGroups(ctx context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {
	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	var nodeGroups []*protos.NodeGroup
	for _, ng := range cluster.NodeGroups {
		nodeGroups = append(nodeGroups, &protos.NodeGroup{
			Id:      ng.Name,
			MinSize: int32(ng.MinSize),
			MaxSize: int32(ng.MaxSize),
			Debug:   fmt.Sprintf("NodeGroup %s for cluster %s", ng.Name, s.clusterID),
		})
	}

	return &protos.NodeGroupsResponse{
		NodeGroups: nodeGroups,
	}, nil
}

func (s *AutoscalerServer) NodeGroupForNode(ctx context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {
	if req.Node == nil || req.Node.ProviderID == "" {
		return &protos.NodeGroupForNodeResponse{}, nil
	}

	// ProviderID format: thecloud://<instance-id>
	providerID := strings.TrimPrefix(req.Node.ProviderID, "thecloud://")
	if providerID == req.Node.ProviderID {
		return &protos.NodeGroupForNodeResponse{}, nil
	}

	// Fetch instance to find its node group metadata
	inst, err := s.client.GetInstanceWithContext(ctx, providerID)
	if err != nil {
		return nil, err
	}

	groupName, ok := inst.Metadata["thecloud.io/node-group"]
	if !ok {
		// Fallback to labels if metadata is missing
		groupName, ok = req.Node.Labels["thecloud.io/node-group"]
		if !ok {
			return &protos.NodeGroupForNodeResponse{}, nil
		}
	}

	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	for _, ng := range cluster.NodeGroups {
		if ng.Name == groupName {
			return &protos.NodeGroupForNodeResponse{
				NodeGroup: &protos.NodeGroup{
					Id:      ng.Name,
					MinSize: int32(ng.MinSize),
					MaxSize: int32(ng.MaxSize),
					Debug:   fmt.Sprintf("NodeGroup %s", ng.Name),
				},
			}, nil
		}
	}

	return &protos.NodeGroupForNodeResponse{}, nil
}

func (s *AutoscalerServer) NodeGroupTargetSize(ctx context.Context, req *protos.NodeGroupTargetSizeRequest) (*protos.NodeGroupTargetSizeResponse, error) {
	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	for _, ng := range cluster.NodeGroups {
		if ng.Name == req.Id {
			return &protos.NodeGroupTargetSizeResponse{
				TargetSize: int32(ng.CurrentSize),
			}, nil
		}
	}

	return nil, fmt.Errorf("node group %s not found", req.Id)
}

func (s *AutoscalerServer) NodeGroupIncreaseSize(ctx context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {
	klog.Infof("NodeGroupIncreaseSize: Request to increase size of node group %s by %d", req.Id, req.Delta)
	
	if req.Delta <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "NodeGroupIncreaseSize: delta must be a positive integer, got %d", req.Delta)
	}

	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	var targetGroup *sdk.NodeGroup
	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == req.Id {
			targetGroup = &cluster.NodeGroups[i]
			break
		}
	}

	if targetGroup == nil {
		return nil, fmt.Errorf("node group %s not found", req.Id)
	}

	newSize := targetGroup.CurrentSize + int(req.Delta)
	if newSize > targetGroup.MaxSize {
		return nil, fmt.Errorf("size would exceed max size")
	}

	_, err = s.client.UpdateNodeGroupWithContext(ctx, s.clusterID, req.Id, sdk.UpdateNodeGroupInput{
		DesiredSize: &newSize,
	})
	if err != nil {
		return nil, err
	}

	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

func (s *AutoscalerServer) NodeGroupDeleteNodes(ctx context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {
	klog.Infof("Request to delete %d nodes from group %s", len(req.Nodes), req.Id)

	successfulTerminations := 0
	for _, node := range req.Nodes {
		providerID := strings.TrimPrefix(node.ProviderID, "thecloud://")
		if providerID == node.ProviderID {
			continue
		}
		if err := s.client.TerminateInstanceWithContext(ctx, providerID); err != nil {
			klog.Errorf("Failed to terminate instance %s: %v", providerID, err)
			return nil, err
		}
		successfulTerminations++
	}

	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	var targetGroup *sdk.NodeGroup
	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == req.Id {
			targetGroup = &cluster.NodeGroups[i]
			break
		}
	}

	if targetGroup != nil {
		newSize := targetGroup.CurrentSize - successfulTerminations
		_, err = s.client.UpdateNodeGroupWithContext(ctx, s.clusterID, req.Id, sdk.UpdateNodeGroupInput{
			DesiredSize: &newSize,
		})
		if err != nil {
			klog.Errorf("Failed to update node group size after node deletion: %v", err)
			return nil, err
		}
	}

	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

func (s *AutoscalerServer) NodeGroupNodes(ctx context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {
	instances, err := s.client.ListInstancesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var protosInstances []*protos.Instance
	for _, inst := range instances {
		if inst.Metadata["thecloud.io/cluster-id"] == s.clusterID &&
			inst.Metadata["thecloud.io/node-group"] == req.Id {
			
			state := protos.InstanceStatus_instanceRunning
			if inst.Status == "PROVISIONING" || inst.Status == "STARTING" {
				state = protos.InstanceStatus_instanceCreating
			} else if inst.Status == "TERMINATING" || inst.Status == "STOPPING" {
				state = protos.InstanceStatus_instanceDeleting
			}

			protosInstances = append(protosInstances, &protos.Instance{
				Id: "thecloud://" + inst.ID,
				Status: &protos.InstanceStatus{
					InstanceState: state,
				},
			})
		}
	}

	return &protosInstancesResponse{
		Instances: protosInstances,
	}, nil
}

type protosInstancesResponse struct {
	Instances []*protos.Instance
}

func (s *AutoscalerServer) GPULabel(ctx context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	return &protos.GPULabelResponse{Label: "thecloud.io/gpu"}, nil
}

func (s *AutoscalerServer) GetAvailableGPUTypes(ctx context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {
	return &protos.GetAvailableGPUTypesResponse{}, nil
}

func (s *AutoscalerServer) Refresh(ctx context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	return &protos.RefreshResponse{}, nil
}

func (s *AutoscalerServer) Cleanup(ctx context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	return &protos.CleanupResponse{}, nil
}

func (s *AutoscalerServer) NodeGroupDecreaseTargetSize(ctx context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	klog.Infof("Request to decrease target size of node group %s by %d", req.Id, req.Delta)
	// Cluster Autoscaler delta is negative
	if req.Delta >= 0 {
		return nil, fmt.Errorf("delta must be negative")
	}

	cluster, err := s.client.GetClusterWithContext(ctx, s.clusterID)
	if err != nil {
		return nil, err
	}

	var targetGroup *sdk.NodeGroup
	for i := range cluster.NodeGroups {
		if cluster.NodeGroups[i].Name == req.Id {
			targetGroup = &cluster.NodeGroups[i]
			break
		}
	}

	if targetGroup == nil {
		return nil, fmt.Errorf("node group %s not found", req.Id)
	}

	newSize := targetGroup.CurrentSize + int(req.Delta)
	if newSize < targetGroup.MinSize {
		return nil, fmt.Errorf("size would be less than min size")
	}

	_, err = s.client.UpdateNodeGroupWithContext(ctx, s.clusterID, req.Id, sdk.UpdateNodeGroupInput{
		DesiredSize: &newSize,
	})
	if err != nil {
		return nil, err
	}

	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

func (s *AutoscalerServer) NodeGroupTemplateNodeInfo(ctx context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGroupTemplateNodeInfo method not implemented")
}

func (s *AutoscalerServer) NodeGroupGetOptions(ctx context.Context, req *protos.NodeGroupAutoscalingOptionsRequest) (*protos.NodeGroupAutoscalingOptionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGroupGetOptions method not implemented")
}

func (s *AutoscalerServer) PricingNodePrice(ctx context.Context, req *protos.PricingNodePriceRequest) (*protos.PricingNodePriceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "PricingNodePrice method not implemented")
}

func (s *AutoscalerServer) PricingPodPrice(ctx context.Context, req *protos.PricingPodPriceRequest) (*protos.PricingPodPriceResponse, error) {
	return nil, status.Error(codes.Unimplemented, "PricingPodPrice method not implemented")
}
