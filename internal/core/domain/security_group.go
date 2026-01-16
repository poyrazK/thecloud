// Package domain defines core business entities.
package domain

import (
	"errors"
	"fmt"
	"net"
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

// Validate checks if the security group fields are valid.
func (sg *SecurityGroup) Validate() error {
	if sg.Name == "" {
		return errors.New("security group name cannot be empty")
	}
	if sg.VPCID == uuid.Nil {
		return errors.New("security group must be associated with a VPC")
	}
	if sg.UserID == uuid.Nil {
		return errors.New("security group must have a user owner")
	}
	return nil
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

// Validate checks if the security rule fields are valid.
func (sr *SecurityRule) Validate() error {
	if err := sr.validateDirection(); err != nil {
		return err
	}
	if err := sr.validateProtocol(); err != nil {
		return err
	}
	if err := sr.validatePorts(); err != nil {
		return err
	}
	if err := sr.validateCIDR(); err != nil {
		return err
	}
	return nil
}

func (sr *SecurityRule) validateDirection() error {
	if sr.Direction != RuleIngress && sr.Direction != RuleEgress {
		return fmt.Errorf("invalid rule direction: %s", sr.Direction)
	}
	return nil
}

func (sr *SecurityRule) validateProtocol() error {
	validProtocols := map[string]bool{"tcp": true, "udp": true, "icmp": true, "all": true}
	if !validProtocols[sr.Protocol] {
		return fmt.Errorf("invalid protocol: %s", sr.Protocol)
	}
	return nil
}

func (sr *SecurityRule) validatePorts() error {
	if sr.Protocol == "all" || sr.Protocol == "icmp" {
		return nil
	}
	if sr.PortMin < 1 || sr.PortMin > 65535 {
		return fmt.Errorf("invalid port_min: %d", sr.PortMin)
	}
	if sr.PortMax < 1 || sr.PortMax > 65535 {
		return fmt.Errorf("invalid port_max: %d", sr.PortMax)
	}
	if sr.PortMin > sr.PortMax {
		return fmt.Errorf("port_min (%d) cannot be greater than port_max (%d)", sr.PortMin, sr.PortMax)
	}
	return nil
}

func (sr *SecurityRule) validateCIDR() error {
	if sr.CIDR == "" {
		return errors.New("CIDR is required")
	}
	_, _, err := net.ParseCIDR(sr.CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}
	return nil
}

// RuleDirection specifies whether a rule applies to incoming or outgoing traffic.
type RuleDirection string

const (
	// RuleIngress applies to incoming traffic to the resource.
	RuleIngress RuleDirection = "ingress"
	// RuleEgress applies to outgoing traffic from the resource.
	RuleEgress RuleDirection = "egress"
)
