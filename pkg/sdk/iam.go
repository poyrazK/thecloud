package sdk

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// CreatePolicy creates a new IAM policy.
func (c *Client) CreatePolicy(ctx context.Context, policy *domain.Policy) (*domain.Policy, error) {
	var res Response[domain.Policy]
	if err := c.postWithContext(ctx, "/iam/policies", policy, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// GetPolicy retrieves a specific IAM policy.
func (c *Client) GetPolicy(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	var res Response[domain.Policy]
	if err := c.getWithContext(ctx, fmt.Sprintf("/iam/policies/%s", id), &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// ListPolicies lists all IAM policies.
func (c *Client) ListPolicies(ctx context.Context) ([]domain.Policy, error) {
	var res Response[[]domain.Policy]
	if err := c.getWithContext(ctx, "/iam/policies", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// UpdatePolicy updates an existing IAM policy.
func (c *Client) UpdatePolicy(ctx context.Context, policy *domain.Policy) (*domain.Policy, error) {
	if policy == nil {
		return nil, errors.New(errors.InvalidInput, "policy cannot be nil")
	}
	var res Response[domain.Policy]
	if err := c.putWithContext(ctx, fmt.Sprintf("/iam/policies/%s", policy.ID), policy, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// DeletePolicy removes an IAM policy.
func (c *Client) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	return c.deleteWithContext(ctx, fmt.Sprintf("/iam/policies/%s", id), nil)
}

// AttachPolicyToUser assigns a policy to a user.
func (c *Client) AttachPolicyToUser(ctx context.Context, userID, policyID uuid.UUID) error {
	return c.postWithContext(ctx, fmt.Sprintf("/iam/users/%s/policies/%s", userID, policyID), nil, nil)
}

// DetachPolicyFromUser removes a policy assignment from a user.
func (c *Client) DetachPolicyFromUser(ctx context.Context, userID, policyID uuid.UUID) error {
	return c.deleteWithContext(ctx, fmt.Sprintf("/iam/users/%s/policies/%s", userID, policyID), nil)
}

// GetUserPolicies lists all policies assigned to a user.
func (c *Client) GetUserPolicies(ctx context.Context, userID uuid.UUID) ([]domain.Policy, error) {
	var res Response[[]domain.Policy]
	if err := c.getWithContext(ctx, fmt.Sprintf("/iam/users/%s/policies", userID), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
