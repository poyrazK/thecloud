// Package ports defines service and repository interfaces.
package ports

// CreateInstanceOptions encapsulates the requirements for provisioning a new compute resource.
type CreateInstanceOptions struct {
	Name        string            // Friendly name for the instance
	ImageName   string            // Template or image to use (e.g., "ubuntu:latest")
	Ports       []string          // List of ports to expose (e.g., ["80/tcp", "443/tcp"])
	NetworkID   string            // ID of the VPC/Network to join
	VolumeBinds []string          // Storage mappings (e.g., ["/host/path:/container/path"])
	Env         []string          // Environment variables (e.g., ["KEY=VALUE"])
	Cmd         []string          // Optional override command for the instance entrypoint
	Metadata    map[string]string // Key-value metadata for the instance
	Labels      map[string]string // Scheduling or grouping labels
	CPULimit    int64             // CPU cores (or millicores)
	MemoryLimit int64             // Memory in bytes
	DiskLimit   int64             // Disk in bytes
	UserData    string            // Cloud-init user data
}
