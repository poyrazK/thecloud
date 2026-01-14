// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

func (c *Client) CreateStack(name, template string, parameters map[string]string) (*domain.Stack, error) {
	var stack domain.Stack
	err := c.post("/iac/stacks", map[string]interface{}{
		"name":       name,
		"template":   template,
		"parameters": parameters,
	}, &stack)
	if err != nil {
		return nil, err
	}
	return &stack, nil
}

func (c *Client) ListStacks() ([]*domain.Stack, error) {
	var stacks []*domain.Stack
	err := c.get("/iac/stacks", &stacks)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (c *Client) GetStack(id string) (*domain.Stack, error) {
	var stack domain.Stack
	err := c.get(fmt.Sprintf("/iac/stacks/%s", id), &stack)
	if err != nil {
		return nil, err
	}
	return &stack, nil
}

func (c *Client) DeleteStack(id string) error {
	return c.delete(fmt.Sprintf("/iac/stacks/%s", id), nil)
}

func (c *Client) ValidateTemplate(template string) (*domain.TemplateValidateResponse, error) {
	var valResp domain.TemplateValidateResponse
	err := c.post("/iac/validate", map[string]interface{}{
		"template": template,
	}, &valResp)
	if err != nil {
		return nil, err
	}
	return &valResp, nil
}
