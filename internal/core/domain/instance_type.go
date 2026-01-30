package domain

// InstanceType represents a pre-defined resource configuration
type InstanceType struct {
	ID          string  `json:"id"`           // e.g., "basic-1"
	Name        string  `json:"name"`         // Human-friendly name
	VCPUs       int     `json:"vcpus"`        // Number of virtual CPUs
	MemoryMB    int     `json:"memory_mb"`    // Memory in MB
	DiskGB      int     `json:"disk_gb"`      // Root disk size
	NetworkMbps int     `json:"network_mbps"` // Network bandwidth
	PricePerHr  float64 `json:"price_per_hour"`
	Category    string  `json:"category"` // basic, standard, performance, gpu
}
