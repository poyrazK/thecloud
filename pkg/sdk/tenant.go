// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
	"time"
)

// Tenant describes a tenant organization.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListTenants returns all tenants the current user belongs to.
func (c *Client) ListTenants() ([]Tenant, error) {
	var res Response[[]Tenant]
	if err := c.get("/tenants", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// CreateTenant creates a new tenant organization.
func (c *Client) CreateTenant(name, slug string) (*Tenant, error) {
	body := map[string]string{
		"name": name,
		"slug": slug,
	}
	var res Response[Tenant]
	if err := c.post("/tenants", body, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// SwitchTenant changes the default tenant for the user.
func (c *Client) SwitchTenant(id string) error {
	return c.post(fmt.Sprintf("/tenants/%s/switch", id), nil, nil)
}
