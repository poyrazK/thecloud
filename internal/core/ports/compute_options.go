// Package ports defines service and repository interfaces.
package ports

// CreateInstanceOptions encapsulates the requirements for provisioning a new compute resource.
type CreateInstanceOptions struct {
	Name        string            `json:"name"`         // Friendly name for the instance
	ImageName   string            `json:"image_name"`   // Template or image to use (e.g., "ubuntu:latest")
	Ports       []string          `json:"ports"`       // List of ports to expose (e.g., ["80/tcp", "443/tcp"])
	NetworkID   string            `json:"network_id"`   // ID of the VPC/Network to join
	VolumeBinds []string          `json:"volume_binds"` // Storage mappings (e.g., ["/host/path:/container/path"])
	Env         []string          `json:"env"`          // Environment variables (e.g., ["KEY=VALUE"])
	Cmd         []string          `json:"cmd"`          // Optional override command for the instance entrypoint
	Metadata    map[string]string `json:"metadata,omitempty"` // Key-value metadata for the instance
	Labels      map[string]string `json:"labels,omitempty"`   // Scheduling or grouping labels
	CPULimit    int64             `json:"cpu_limit"`    // CPU cores (or millicores)
	MemoryLimit int64             `json:"memory_limit"` // Memory in bytes
	DiskLimit   int64             `json:"disk_limit"`   // Disk in bytes
	UserData    string            `json:"user_data"`    // Cloud-init user data
}
