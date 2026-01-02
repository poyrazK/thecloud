package domain

import (
	"time"

	"github.com/google/uuid"
)

type LBStatus string

const (
	LBStatusCreating LBStatus = "CREATING"
	LBStatusActive   LBStatus = "ACTIVE"
	LBStatusDraining LBStatus = "DRAINING"
	LBStatusDeleted  LBStatus = "DELETED"
)

type LoadBalancer struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	Name           string    `json:"name"`
	VpcID          uuid.UUID `json:"vpc_id"`
	Port           int       `json:"port"`
	Algorithm      string    `json:"algorithm"` // "round-robin" | "least-conn"
	Status         LBStatus  `json:"status"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
}

type LBTarget struct {
	ID         uuid.UUID `json:"id"`
	LBID       uuid.UUID `json:"lb_id"`
	InstanceID uuid.UUID `json:"instance_id"`
	Port       int       `json:"port"`
	Weight     int       `json:"weight"`
	Health     string    `json:"health"` // "healthy" | "unhealthy" | "unknown"
}

type HealthCheckConfig struct {
	Path               string `json:"path"`
	IntervalSeconds    int    `json:"interval_seconds"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
}
