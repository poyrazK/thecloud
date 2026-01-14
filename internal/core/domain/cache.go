// Package domain contains the core domain models for the cloud platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// CacheEngine represents the type of caching technology.
type CacheEngine string

const (
	// EngineRedis represents a Redis-based cache instance.
	EngineRedis CacheEngine = "redis"
)

// CacheStatus represents the lifecycle state of a cache instance.
type CacheStatus string

const (
	// CacheStatusCreating indicates the cache is being provisioned.
	CacheStatusCreating CacheStatus = "CREATING"
	// CacheStatusRunning indicates the cache is fully operational.
	CacheStatusRunning CacheStatus = "RUNNING"
	// CacheStatusStopped indicates the cache instance is halted.
	CacheStatusStopped CacheStatus = "STOPPED"
	// CacheStatusDeleting indicates the cache is being removed.
	CacheStatusDeleting CacheStatus = "DELETING"
	// CacheStatusFailed indicates the cache encountered an error.
	CacheStatusFailed CacheStatus = "FAILED"
)

// Cache represents a managed caching service instance (e.g. Redis).
type Cache struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Name        string      `json:"name"`
	Engine      CacheEngine `json:"engine"`
	Version     string      `json:"version"` // Engine version (e.g. "7.0")
	Status      CacheStatus `json:"status"`
	VpcID       *uuid.UUID  `json:"vpc_id,omitempty"` // Optional private networking
	ContainerID string      `json:"container_id,omitempty"`
	Port        int         `json:"port"`
	Password    string      `json:"-"` // Never serialize password to JSON
	MemoryMB    int         `json:"memory_mb"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
