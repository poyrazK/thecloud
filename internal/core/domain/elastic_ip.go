// Package domain defines core business entities.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ElasticIPStatus represents the lifecycle state of an Elastic IP.
type ElasticIPStatus string

const (
	// EIPStatusAllocated indicates the IP is reserved by a user but not attached to any resource.
	EIPStatusAllocated ElasticIPStatus = "allocated"

	// EIPStatusAssociated indicates the IP is currently mapped to a compute instance.
	EIPStatusAssociated ElasticIPStatus = "associated"

	// EIPStatusReleased indicates the IP has been returned to the pool (soft delete state).
	EIPStatusReleased ElasticIPStatus = "released"
)

// ElasticIP represents a static public IP address that can be remapped between instances.
type ElasticIP struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	TenantID   uuid.UUID       `json:"tenant_id"`
	PublicIP   string          `json:"public_ip"`
	InstanceID *uuid.UUID      `json:"instance_id,omitempty"`
	VpcID      *uuid.UUID      `json:"vpc_id,omitempty"`
	Status     ElasticIPStatus `json:"status"`
	ARN        string          `json:"arn"` // arn:thecloud:vpc:{region}:{user}:eip/{id}
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Validate checks if the Elastic IP model is valid.
func (e *ElasticIP) Validate() error {
	if e.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}
	if e.TenantID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	if e.PublicIP == "" {
		return errors.New("public IP address is required")
	}
	if e.Status == "" {
		return errors.New("status is required")
	}
	return nil
}
