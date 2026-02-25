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
func (c *Client) CreateKey(name string) (string, error) {
	body := map[string]string{"name": name}
	var res Response[struct {
		Key string `json:"key"`
	}]

	resp, err := c.resty.R().
		SetBody(body).
		SetResult(&res).
		Post(c.apiURL + "/auth/keys")

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("api error: %s", resp.String())
	}
	return res.Data.Key, nil
}
