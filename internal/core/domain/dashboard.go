// Package domain defines core business entities.
package domain

import "time"

// ResourceSummary provides a high-level overview of resource health and counts across the platform.
type ResourceSummary struct {
	TotalInstances   int `json:"total_instances"`   // All compute instances regardless of state
	RunningInstances int `json:"running_instances"` // Instances currently in a RUNNING state
	StoppedInstances int `json:"stopped_instances"` // Instances currently in a STOPPED state
	TotalVolumes     int `json:"total_volumes"`     // Total number of block storage volumes
	AttachedVolumes  int `json:"attached_volumes"`  // Volumes currently mounted to instances
	TotalVPCs        int `json:"total_vpcs"`        // Total number of virtual private clouds
	TotalStorageMB   int `json:"total_storage_mb"`  // Aggregate storage allocated across all volumes
}

// MetricPoint represents a single quantitative measurement at a specific point in time.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"` // The measured value (e.g., percentage or byte count)
	Label     string    `json:"label"` // Optional category or identifier (e.g., "avg", "max")
}

// DashboardStats aggregates various metrics and logs to populate the user dashboard.
type DashboardStats struct {
	Summary       ResourceSummary `json:"summary"`
	RecentEvents  []Event         `json:"recent_events"`  // Latest audit/system events
	CPUHistory    []MetricPoint   `json:"cpu_history"`    // Normalized CPU utilization over time
	MemoryHistory []MetricPoint   `json:"memory_history"` // Normalized memory utilization over time
}
