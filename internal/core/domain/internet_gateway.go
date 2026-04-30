package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// IGWStatus represents the attachment state of an Internet Gateway.
type IGWStatus string

const (
	IGWStatusDetached IGWStatus = "detached"
	IGWStatusAttached IGWStatus = "attached"
)

// InternetGateway provides a path for traffic between a VPC and the internet.
// An IGW can only be attached to one VPC at a time.
type InternetGateway struct {
	ID        uuid.UUID  `json:"id"`
	VPCID     *uuid.UUID `json:"vpc_id,omitempty"`
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Status    IGWStatus  `json:"status"`
	ARN       string     `json:"arn"`
	CreatedAt time.Time  `json:"created_at"`
}

// Validate checks if the internet gateway fields are valid.
func (igw *InternetGateway) Validate() error {
	if igw.UserID == uuid.Nil {
		return errors.New("internet gateway must have a user owner")
	}
	if igw.TenantID == uuid.Nil {
		return errors.New("internet gateway must have a tenant")
	}
	if !isValidIGWStatus(igw.Status) {
		return errors.New("invalid IGW status")
	}
	return nil
}

func isValidIGWStatus(s IGWStatus) bool {
	switch s {
	case IGWStatusDetached, IGWStatusAttached:
		return true
	}
	return false
}

// CanAttach checks if the IGW can be attached to a VPC.
func (igw *InternetGateway) CanAttach() bool {
	return igw.Status == IGWStatusDetached && igw.VPCID == nil
}

// CanDetach checks if the IGW can be detached from its VPC.
func (igw *InternetGateway) CanDetach() bool {
	return igw.Status == IGWStatusAttached
}

// IsAttached checks if the IGW is currently attached.
func (igw *InternetGateway) IsAttached() bool {
	return igw.Status == IGWStatusAttached && igw.VPCID != nil
}
