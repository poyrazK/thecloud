// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"time"

	"github.com/google/uuid"
)

const clustersPath = "/clusters"

// Cluster represents a managed Kubernetes cluster in the SDK.
type Cluster struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	UserID             uuid.UUID `json:"user_id"`
	VpcID              uuid.UUID `json:"vpc_id"`
	Version            string    `json:"version"`
	ControlPlaneIPs    []string  `json:"control_plane_ips"`
	WorkerCount        int       `json:"worker_count"`
	HAEnabled          bool      `json:"ha_enabled"`
	APIServerLBAddress *string   `json:"api_server_lb_address,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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
	var resp Response[*Cluster]
	if err := c.get(clustersPath+"/"+id, &resp); err != nil {
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
