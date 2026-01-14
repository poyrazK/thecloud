// Package ports defines service and repository interfaces.
package ports

// CreateInstanceOptions encapsulates the requirements for provisioning a new compute resource.
type CreateInstanceOptions struct {
	Name        string   // Friendly name for the instance
	ImageName   string   // Template or image to use (e.g., "ubuntu:latest")
	Ports       []string // List of ports to expose (e.g., ["80/tcp", "443/tcp"])
	NetworkID   string   // ID of the VPC/Network to join
	VolumeBinds []string // Storage mappings (e.g., ["/host/path:/container/path"])
	Env         []string // Environment variables (e.g., ["KEY=VALUE"])
	Cmd         []string // Optional override command for the instance entrypoint
}
