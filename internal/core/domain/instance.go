// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// InstanceStatus represents the current state of a compute instance.
//
// Instances transition through these states during their lifecycle:
//
//	STARTING → RUNNING → STOPPED → DELETED
//
// If an error occurs during any transition, the instance moves to ERROR state.
type InstanceStatus string

const (
	// StatusStarting indicates the instance is being created and initialized.
	// This is a transient state during instance launch.
	StatusStarting InstanceStatus = "STARTING"

	// StatusRunning indicates the instance is active and operational.
	// The underlying container/VM is running and accepting connections.
	StatusRunning InstanceStatus = "RUNNING"

	// StatusStopped indicates the instance has been stopped.
	// Resources are still allocated but the instance is not running.
	StatusStopped InstanceStatus = "STOPPED"

	// StatusError indicates an error occurred during instance operations.
	// Check logs for details. Manual intervention may be required.
	StatusError InstanceStatus = "ERROR"

	// StatusDeleted indicates the instance has been permanently removed.
	// All associated resources have been cleaned up.
	StatusDeleted InstanceStatus = "DELETED"
)

const (
	// MinPort is the minimum valid port number (0 means auto-assign).
	MinPort = 0

	// MaxPort is the maximum valid port number (65535).
	MaxPort = 65535

	// MaxPortsPerInstance limits the number of port mappings per instance
	// to prevent resource exhaustion.
	MaxPortsPerInstance = 10
)

// Instance represents a compute instance (container or VM).
//
// Instances run container images or VM templates with optional port mappings
// and can be attached to VPCs for advanced networking. The Version field
// enables optimistic locking to prevent conflicting updates.
//
// Port format: "hostPort:containerPort" (e.g., "8080:80,443:443")
type Instance struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`                // Owner for multi-tenancy
	Name        string         `json:"name"`                   // Unique per user
	Image       string         `json:"image"`                  // Container/VM image
	ContainerID string         `json:"container_id,omitempty"` // Backend identifier
	Status      InstanceStatus `json:"status"`
	Ports       string         `json:"ports,omitempty"`  // "host:container" mappings
	VpcID       *uuid.UUID     `json:"vpc_id,omitempty"` // Optional VPC attachment
	SubnetID    *uuid.UUID     `json:"subnet_id,omitempty"`
	PrivateIP   string         `json:"private_ip,omitempty"` // VPC private IP
	OvsPort     string         `json:"ovs_port,omitempty"`   // OVS port name
	Version     int            `json:"version"`              // Optimistic locking
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// InstanceStats contains real-time resource usage metrics.
// Values are instantaneous snapshots from the compute backend.
type InstanceStats struct {
	CPUPercentage    float64 `json:"cpu_percentage"`
	MemoryUsageBytes float64 `json:"memory_usage_bytes"`
	MemoryLimitBytes float64 `json:"memory_limit_bytes"`
	MemoryPercentage float64 `json:"memory_percentage"`
}

// RawDockerStats mirrors Docker's stats payload for CPU/memory calculations.
type RawDockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
}
