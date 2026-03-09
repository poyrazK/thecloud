// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"context"
	"net/url"
	"time"

	"github.com/google/uuid"
)

const clustersPath = "/clusters"

// NodeGroup represents a pool of similar worker nodes in a cluster.
type NodeGroup struct {
	ID           uuid.UUID `json:"id"`
	ClusterID    uuid.UUID `json:"cluster_id"`
	Name         string    `json:"name"`
	InstanceType string    `json:"instance_type"`
	MinSize      int       `json:"min_size"`
	MaxSize      int       `json:"max_size"`
	CurrentSize  int       `json:"current_size"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Cluster represents a managed Kubernetes cluster in the SDK.
type Cluster struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	UserID             uuid.UUID   `json:"user_id"`
	VpcID              uuid.UUID   `json:"vpc_id"`
	Version            string      `json:"version"`
	ControlPlaneIPs    []string    `json:"control_plane_ips"`
	WorkerCount        int         `json:"worker_count"`
	HAEnabled          bool        `json:"ha_enabled"`
	APIServerLBAddress *string     `json:"api_server_lb_address,omitempty"`
	Status             string      `json:"status"`
	NodeGroups         []NodeGroup `json:"node_groups,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// CreateClusterInput defines the input for creating a cluster.
type CreateClusterInput struct {
	Name             string    `json:"name"`
	VpcID            uuid.UUID `json:"vpc_id"`
	Version          string    `json:"version"`
	WorkerCount      int       `json:"workers"`
	NetworkIsolation bool      `json:"network_isolation"`
	HA               bool      `json:"ha"`
}

// NodeGroupInput defines the input for adding a node group.
type NodeGroupInput struct {
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
	MinSize      int    `json:"min_size"`
	MaxSize      int    `json:"max_size"`
	DesiredSize  int    `json:"desired_size"`
}

// UpdateNodeGroupInput defines the input for updating a node group.
type UpdateNodeGroupInput struct {
	DesiredSize *int `json:"desired_size"`
	MinSize     *int `json:"min_size"`
	MaxSize     *int `json:"max_size"`
}

// ClusterHealth represents the operational status of a cluster.
type ClusterHealth struct {
	Status     string `json:"status"`
	APIServer  bool   `json:"api_server"`
	NodesTotal int    `json:"nodes_total"`
	NodesReady int    `json:"nodes_ready"`
	Message    string `json:"message,omitempty"`
}

// ScaleClusterInput defines the input for scaling worker nodes.
type ScaleClusterInput struct {
	Workers int `json:"workers"`
}

// ListClusters returns all clusters for the current user.
func (c *Client) ListClusters() ([]*Cluster, error) {
	var resp Response[[]*Cluster]
	if err := c.get(clustersPath, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// CreateCluster initiates cluster provisioning.
func (c *Client) CreateCluster(input *CreateClusterInput) (*Cluster, error) {
	var resp Response[*Cluster]
	if err := c.post(clustersPath, input, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetCluster retrieves cluster details by ID.
func (c *Client) GetCluster(id string) (*Cluster, error) {
	return c.GetClusterWithContext(context.Background(), id)
}

// GetClusterWithContext retrieves cluster details by ID with context support.
func (c *Client) GetClusterWithContext(ctx context.Context, id string) (*Cluster, error) {
	var resp Response[*Cluster]
	if err := c.getWithContext(ctx, clustersPath+"/"+id, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// DeleteCluster removes a cluster.
func (c *Client) DeleteCluster(id string) error {
	var resp Response[any]
	return c.delete(clustersPath+"/"+id, &resp)
}

// GetKubeconfig retrieves the cluster kubeconfig, optionally for a specific role.
func (c *Client) GetKubeconfig(id string, role string) (string, error) {
	path := clustersPath + "/" + id + "/kubeconfig"
	if role != "" {
		path += "?role=" + role
	}
	var resp Response[string]
	if err := c.get(path, &resp); err != nil {
		return "", err
	}
	return resp.Data, nil
}

// RepairCluster triggers a re-run of critical provisioning steps.
func (c *Client) RepairCluster(id string) error {
	var resp Response[any]
	return c.post(clustersPath+"/"+id+"/repair", nil, &resp)
}

// ScaleCluster adjusts the number of worker nodes.
func (c *Client) ScaleCluster(id string, workers int) error {
	var resp Response[any]
	input := &ScaleClusterInput{Workers: workers}
	return c.post(clustersPath+"/"+id+"/scale", input, &resp)
}

// GetClusterHealth retrieved the operational health of the cluster.
func (c *Client) GetClusterHealth(id string) (*ClusterHealth, error) {
	var resp Response[*ClusterHealth]
	if err := c.get(clustersPath+"/"+id+"/health", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpgradeClusterInput defines the input for upgrading a cluster.
type UpgradeClusterInput struct {
	Version string `json:"version"`
}

// UpgradeCluster initiates an asynchronous version upgrade.
func (c *Client) UpgradeCluster(id string, version string) error {
	var resp Response[any]
	input := &UpgradeClusterInput{Version: version}
	return c.post(clustersPath+"/"+id+"/upgrade", input, &resp)
}

// RotateSecrets triggers a renewal of cluster certificates.
func (c *Client) RotateSecrets(id string) error {
	var resp Response[any]
	return c.post(clustersPath+"/"+id+"/rotate-secrets", nil, &resp)
}

// CreateBackup initiates a cluster state snapshot.
func (c *Client) CreateBackup(id string) error {
	var resp Response[any]
	return c.post(clustersPath+"/"+id+"/backups", nil, &resp)
}

// RestoreBackupInput defines the input for restoring a cluster from backup.
type RestoreBackupInput struct {
	BackupPath string `json:"backup_path"`
}

// RestoreBackup initiates a cluster restoration from a specific path.
func (c *Client) RestoreBackup(id string, backupPath string) error {
	var resp Response[any]
	input := &RestoreBackupInput{BackupPath: backupPath}
	return c.post(clustersPath+"/"+id+"/restore", input, &resp)
}

// AddNodeGroup adds a new node pool to the cluster.
func (c *Client) AddNodeGroup(clusterID string, input NodeGroupInput) (*NodeGroup, error) {
	var resp Response[*NodeGroup]
	if err := c.post(clustersPath+"/"+clusterID+"/nodegroups", input, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpdateNodeGroupWithContext updates a node group's parameters with context support.
func (c *Client) UpdateNodeGroupWithContext(ctx context.Context, clusterID string, name string, input UpdateNodeGroupInput) (*NodeGroup, error) {
	var resp Response[*NodeGroup]
	if err := c.putWithContext(ctx, clustersPath+"/"+clusterID+"/nodegroups/"+url.PathEscape(name), input, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// DeleteNodeGroup removes a node group.
func (c *Client) DeleteNodeGroup(clusterID string, name string) error {
	var resp Response[any]
	return c.delete(clustersPath+"/"+clusterID+"/nodegroups/"+url.PathEscape(name), &resp)
}
