package sdk

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// CreatePolicy creates a new IAM policy.
func (c *Client) CreatePolicy(policy *domain.Policy) (*domain.Policy, error) {
	var res Response[domain.Policy]
	if err := c.post("/iam/policies", policy, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// GetPolicy retrieves a specific IAM policy.
func (c *Client) GetPolicy(id uuid.UUID) (*domain.Policy, error) {
	var res Response[domain.Policy]
	if err := c.get(fmt.Sprintf("/iam/policies/%s", id), &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// ListPolicies lists all IAM policies.
func (c *Client) ListPolicies() ([]domain.Policy, error) {
	var res Response[[]domain.Policy]
	if err := c.get("/iam/policies", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// UpdatePolicy updates an existing IAM policy.
func (c *Client) UpdatePolicy(policy *domain.Policy) (*domain.Policy, error) {
	var res Response[domain.Policy]
	if err := c.put(fmt.Sprintf("/iam/policies/%s", policy.ID), policy, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// DeletePolicy removes an IAM policy.
func (c *Client) DeletePolicy(id uuid.UUID) error {
	return c.delete(fmt.Sprintf("/iam/policies/%s", id), nil)
}

// AttachPolicyToUser assigns a policy to a user.
func (c *Client) AttachPolicyToUser(userID, policyID uuid.UUID) error {
	return c.post(fmt.Sprintf("/iam/users/%s/policies/%s", userID, policyID), nil, nil)
}

// DetachPolicyFromUser removes a policy assignment from a user.
func (c *Client) DetachPolicyFromUser(userID, policyID uuid.UUID) error {
	return c.delete(fmt.Sprintf("/iam/users/%s/policies/%s", userID, policyID), nil)
}

// GetUserPolicies lists all policies assigned to a user.
func (c *Client) GetUserPolicies(userID uuid.UUID) ([]domain.Policy, error) {
	var res Response[[]domain.Policy]
	if err := c.get(fmt.Sprintf("/iam/users/%s/policies", userID), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
