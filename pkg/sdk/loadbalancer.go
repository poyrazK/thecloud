// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"context"
	"fmt"
)

// LBStatus represents the lifecycle state of a load balancer.
type LBStatus string

// LoadBalancer describes a load balancer resource.
type LoadBalancer struct {
	ID             string   `json:"id"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
	Name           string   `json:"name"`
	VpcID          string   `json:"vpc_id"`
	Port           int      `json:"port"`
	Algorithm      string   `json:"algorithm"`
	Status         LBStatus `json:"status"`
}

// LBTarget describes a load balancer target.
type LBTarget struct {
	ID         string `json:"id"`
	LBID       string `json:"lb_id"`
	InstanceID string `json:"instance_id"`
	Port       int    `json:"port"`
	Weight     int    `json:"weight"`
	Health     string `json:"health"`
}

func (c *Client) CreateLB(name, vpcID string, port int, algo string) (*LoadBalancer, error) {
	return c.CreateLBWithContext(context.Background(), name, vpcID, port, algo)
}

func (c *Client) CreateLBWithContext(ctx context.Context, name, vpcID string, port int, algo string) (*LoadBalancer, error) {
	req := map[string]interface{}{
		"name":      name,
		"vpc_id":    vpcID,
		"port":      port,
		"algorithm": algo,
	}

	var resp Response[LoadBalancer]
	if err := c.postWithContext(ctx, "/lb", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) ListLBs() ([]LoadBalancer, error) {
	return c.ListLBsWithContext(context.Background())
}

func (c *Client) ListLBsWithContext(ctx context.Context) ([]LoadBalancer, error) {
	var resp Response[[]LoadBalancer]
	if err := c.getWithContext(ctx, "/lb", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetLB(id string) (*LoadBalancer, error) {
	var resp Response[LoadBalancer]
	if err := c.get(fmt.Sprintf("/lb/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteLB(id string) error {
	return c.DeleteLBWithContext(context.Background(), id)
}

func (c *Client) DeleteLBWithContext(ctx context.Context, id string) error {
	return c.deleteWithContext(ctx, fmt.Sprintf("/lb/%s", id), nil)
}

func (c *Client) AddLBTarget(lbID, instanceID string, port, weight int) error {
	return c.AddLBTargetWithContext(context.Background(), lbID, instanceID, port, weight)
}

func (c *Client) AddLBTargetWithContext(ctx context.Context, lbID, instanceID string, port, weight int) error {
	req := map[string]interface{}{
		"instance_id": instanceID,
		"port":        port,
		"weight":      weight,
	}

	return c.postWithContext(ctx, fmt.Sprintf("/lb/%s/targets", lbID), req, nil)
}

func (c *Client) RemoveLBTarget(lbID, instanceID string) error {
	return c.RemoveLBTargetWithContext(context.Background(), lbID, instanceID)
}

func (c *Client) RemoveLBTargetWithContext(ctx context.Context, lbID, instanceID string) error {
	return c.deleteWithContext(ctx, fmt.Sprintf("/lb/%s/targets/%s", lbID, instanceID), nil)
}

func (c *Client) ListLBTargets(lbID string) ([]LBTarget, error) {
	return c.ListLBTargetsWithContext(context.Background(), lbID)
}

func (c *Client) ListLBTargetsWithContext(ctx context.Context, lbID string) ([]LBTarget, error) {
	var resp Response[[]LBTarget]
	if err := c.getWithContext(ctx, fmt.Sprintf("/lb/%s/targets", lbID), &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
