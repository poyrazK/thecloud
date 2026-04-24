// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
	"time"
)

// NATGatewayStatus represents the state of a NAT Gateway.
type NATGatewayStatus string

const (
	NATGatewayStatusPending NATGatewayStatus = "pending"
	NATGatewayStatusActive  NATGatewayStatus = "active"
	NATGatewayStatusFailed  NATGatewayStatus = "failed"
	NATGatewayStatusDeleted NATGatewayStatus = "deleted"
)

// NATGateway describes a NAT gateway resource.
type NATGateway struct {
	ID          string          `json:"id"`
	VPCID       string          `json:"vpc_id"`
	SubnetID    string          `json:"subnet_id"`
	ElasticIPID string          `json:"elastic_ip_id"`
	UserID      string          `json:"user_id"`
	TenantID    string          `json:"tenant_id"`
	Status      NATGatewayStatus `json:"status"`
	PrivateIP   string          `json:"private_ip"`
	ARN         string          `json:"arn"`
	CreatedAt   time.Time       `json:"created_at"`
}

// CreateNATGateway creates a new NAT gateway in a subnet with an elastic IP.
func (c *Client) CreateNATGateway(subnetID, eipID string) (*NATGateway, error) {
	body := map[string]string{
		"subnet_id":     subnetID,
		"elastic_ip_id": eipID,
	}
	var res Response[NATGateway]
	if err := c.post("/nat-gateways", body, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// ListNATGateways returns all NAT gateways for a VPC.
func (c *Client) ListNATGateways(vpcID string) ([]NATGateway, error) {
	var res Response[[]NATGateway]
	path := "/nat-gateways"
	if vpcID != "" {
		path = fmt.Sprintf("%s?vpc_id=%s", path, vpcID)
	}
	if err := c.get(path, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// GetNATGateway retrieves details of a specific NAT gateway.
func (c *Client) GetNATGateway(id string) (*NATGateway, error) {
	var res Response[NATGateway]
	if err := c.get(fmt.Sprintf("/nat-gateways/%s", id), &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// DeleteNATGateway permanently removes a NAT gateway.
func (c *Client) DeleteNATGateway(id string) error {
	return c.delete(fmt.Sprintf("/nat-gateways/%s", id), nil)
}