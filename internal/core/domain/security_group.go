// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecurityGroup acts as a virtual firewall for compute instances to control inbound and outbound traffic.
type SecurityGroup struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	VPCID       uuid.UUID      `json:"vpc_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ARN         string         `json:"arn"` // Unique identifier (arn:thecloud:vpc:{region}:{user}:sg/{name})
	Rules       []SecurityRule `json:"rules,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// SecurityRule defines a single traffic filtering criteria.
type SecurityRule struct {
	ID        uuid.UUID     `json:"id"`
	GroupID   uuid.UUID     `json:"group_id"`
	Direction RuleDirection `json:"direction"`
	Protocol  string        `json:"protocol"` // Traffic protocol (e.g., "tcp", "udp", "icmp", "all")
	PortMin   int           `json:"port_min,omitempty"`
	PortMax   int           `json:"port_max,omitempty"`
	CIDR      string        `json:"cidr"`     // Targeted IPv4 range (e.g., "0.0.0.0/0")
	Priority  int           `json:"priority"` // Evaluation order (lower values evaluated first)
	CreatedAt time.Time     `json:"created_at"`
}

// RuleDirection specifies whether a rule applies to incoming or outgoing traffic.
type RuleDirection string

const (
	// RuleIngress applies to incoming traffic to the resource.
	RuleIngress RuleDirection = "ingress"
	// RuleEgress applies to outgoing traffic from the resource.
	RuleEgress RuleDirection = "egress"
)
