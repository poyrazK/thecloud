// Package sdk provides the official Go SDK for the platform.
package sdk

import "fmt"

// CreateKey requests a new API key for the given name.
func (c *Client) CreateKey(name string) (string, error) {
	body := map[string]string{"name": name}
	var res Response[struct {
		Key string `json:"key"`
	}]

	// For bootstrapping, we use a basic client without a key if needed,
	// but CreateKey endpoint usually doesn't require a key or has a different policy.
	// We'll use the existing resty client.

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
