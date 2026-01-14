// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
	"time"
)

// Subnet describes a VPC subnet.
type Subnet struct {
	ID        string    `json:"id"`
	VpcID     string    `json:"vpc_id"`
	Name      string    `json:"name"`
	CIDRBlock string    `json:"cidr_block"`
	AZ        string    `json:"availability_zone"`
	GatewayIP string    `json:"gateway_ip"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Client) ListSubnets(vpcID string) ([]*Subnet, error) {
	var resp Response[[]*Subnet]
	err := c.get(fmt.Sprintf("/vpcs/%s/subnets", vpcID), &resp)
	return resp.Data, err
}

func (c *Client) CreateSubnet(vpcID, name, cidr, az string) (*Subnet, error) {
	var resp Response[*Subnet]
	body := map[string]string{
		"name":              name,
		"cidr_block":        cidr,
		"availability_zone": az,
	}
	err := c.post(fmt.Sprintf("/vpcs/%s/subnets", vpcID), body, &resp)
	return resp.Data, err
}

func (c *Client) DeleteSubnet(id string) error {
	return c.delete(fmt.Sprintf("/subnets/%s", id), nil)
}

func (c *Client) GetSubnet(id string) (*Subnet, error) {
	var resp Response[*Subnet]
	err := c.get(fmt.Sprintf("/subnets/%s", id), &resp)
	return resp.Data, err
}
