package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// NATGatewayStatus represents the state of a NAT Gateway.
type NATGatewayStatus string

const (
	NATGatewayStatusPending NATGatewayStatus = "pending"
	NATGatewayStatusActive  NATGatewayStatus = "active"
	NATGatewayStatusFailed  NATGatewayStatus = "failed"
	NATGatewayStatusDeleted NATGatewayStatus = "deleted"
)

// NATGateway provides SNAT (Source NAT) for instances in private subnets
// to access the internet while preventing inbound connections from the internet.
type NATGateway struct {
	ID          uuid.UUID        `json:"id"`
	VPCID       uuid.UUID        `json:"vpc_id"`
	SubnetID    uuid.UUID        `json:"subnet_id"`
	ElasticIPID uuid.UUID        `json:"elastic_ip_id"`
	UserID      uuid.UUID        `json:"user_id"`
	TenantID    uuid.UUID        `json:"tenant_id"`
	Status      NATGatewayStatus `json:"status"`
	PrivateIP   string           `json:"private_ip"`
	ARN         string           `json:"arn"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Validate checks if the NAT gateway fields are valid.
func (ng *NATGateway) Validate() error {
	if ng.VPCID == uuid.Nil {
		return errors.New("NAT gateway must be associated with a VPC")
	}
	if ng.SubnetID == uuid.Nil {
		return errors.New("NAT gateway must be placed in a subnet")
	}
	if ng.ElasticIPID == uuid.Nil {
		return errors.New("NAT gateway requires an elastic IP")
	}
	if ng.UserID == uuid.Nil {
		return errors.New("NAT gateway must have a user owner")
	}
	if ng.TenantID == uuid.Nil {
		return errors.New("NAT gateway must have a tenant")
	}
	if !isValidNATGatewayStatus(ng.Status) {
		return errors.New("invalid NAT gateway status")
	}
	return nil
}

func isValidNATGatewayStatus(s NATGatewayStatus) bool {
	switch s {
	case NATGatewayStatusPending, NATGatewayStatusActive,
		NATGatewayStatusFailed, NATGatewayStatusDeleted:
		return true
	}
	return false
}

// IsActive checks if the NAT gateway is operational.
func (ng *NATGateway) IsActive() bool {
	return ng.Status == NATGatewayStatusActive
}