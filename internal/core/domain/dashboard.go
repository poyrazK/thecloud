package domain

import "time"

// ResourceSummary provides an overview of all resources in the system.
type ResourceSummary struct {
	TotalInstances   int `json:"total_instances"`
	RunningInstances int `json:"running_instances"`
	StoppedInstances int `json:"stopped_instances"`
	TotalVolumes     int `json:"total_volumes"`
	AttachedVolumes  int `json:"attached_volumes"`
	TotalVPCs        int `json:"total_vpcs"`
	TotalStorageMB   int `json:"total_storage_mb"`
}

// MetricPoint represents a single data point for time-series metrics.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label"`
}

// DashboardStats contains aggregated metrics for the dashboard.
type DashboardStats struct {
	Summary       ResourceSummary `json:"summary"`
	RecentEvents  []Event         `json:"recent_events"`
	CPUHistory    []MetricPoint   `json:"cpu_history"`
	MemoryHistory []MetricPoint   `json:"memory_history"`
}
