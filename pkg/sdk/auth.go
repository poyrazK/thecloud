// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"

	"github.com/poyrazk/thecloud/internal/core/domain"
)

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the authenticated user and API key.
type LoginResponse struct {
	User   *domain.User `json:"user"`
	APIKey string       `json:"api_key"`
}

// Register creates a new user account.
func (c *Client) Register(email, password, name string) (*domain.User, error) {
	req := RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}
	var res Response[*domain.User]
	if err := c.post("/auth/register", req, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// Login authenticates a user and returns an API key.
func (c *Client) Login(email, password string) (*LoginResponse, error) {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}
	var res Response[*LoginResponse]
	if err := c.post("/auth/login", req, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// WhoAmI returns the currently authenticated user's info.
func (c *Client) WhoAmI() (*domain.User, error) {
	var res Response[domain.User]
	if err := c.get("/auth/me", &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// CreateKey requests a new API key for the given name.
func (c *Client) CreateKey(name string) (*domain.APIKey, error) {
	body := map[string]string{"name": name}
	var res Response[*domain.APIKey]
	if err := c.post("/auth/keys", body, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// ListKeys returns all API keys for the current user.
func (c *Client) ListKeys() ([]*domain.APIKey, error) {
	var res Response[[]*domain.APIKey]
	if err := c.get("/auth/keys", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// RevokeKey deletes an API key.
func (c *Client) RevokeKey(id string) error {
	return c.delete(fmt.Sprintf("/auth/keys/%s", id), nil)
}

// RotateKey rotates an API key.
func (c *Client) RotateKey(id string) (*domain.APIKey, error) {
	var res Response[*domain.APIKey]
	if err := c.post(fmt.Sprintf("/auth/keys/%s/rotate", id), nil, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
