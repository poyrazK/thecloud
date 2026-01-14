// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Subnet represents a subdivision of a VPC's IP address range.
// Instances launch into subnets to receive private IP addresses.
type Subnet struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	VPCID            uuid.UUID `json:"vpc_id"`
	Name             string    `json:"name"`
	CIDRBlock        string    `json:"cidr_block"`        // IPv4 range (e.g. "10.0.1.0/24")
	AvailabilityZone string    `json:"availability_zone"` // Physical zone (e.g. "us-east-1a")
	GatewayIP        string    `json:"gateway_ip"`        // Router IP (usually first IP in block)
	ARN              string    `json:"arn"`               // Amazon Resource Name compatible ID
	Status           string    `json:"status"`            // e.g. "AVAILABLE"
	CreatedAt        time.Time `json:"created_at"`
}
