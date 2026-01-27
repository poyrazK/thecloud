// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"bytes"
	"fmt"
	"time"
)

// Function describes a serverless function.
type Function struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`
	Handler   string    `json:"handler"`
	CodePath  string    `json:"code_path"`
	Timeout   int       `json:"timeout"`
	MemoryMB  int       `json:"memory_mb"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Invocation represents a function invocation result.
type Invocation struct {
	ID         string     `json:"id"`
	FunctionID string     `json:"function_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	DurationMs int        `json:"duration_ms"`
	StatusCode int        `json:"status_code"`
	Logs       string     `json:"logs"`
}

const functionsPath = "/functions/"

func (c *Client) CreateFunction(name, runtime, handler string, code []byte) (*Function, error) {
	var resp Response[Function]
	req := c.resty.R().
		SetFileReader("code", "code.zip", bytes.NewReader(code)).
		SetFormData(map[string]string{
			"name":    name,
			"runtime": runtime,
			"handler": handler,
		}).
		SetResult(&resp)

	httpResp, err := req.Post(c.apiURL + "/functions")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("api error: %s", httpResp.String())
	}

	return &resp.Data, nil
}

func (c *Client) ListFunctions() ([]*Function, error) {
	var resp Response[[]*Function]
	if err := c.get("/functions", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetFunction(id string) (*Function, error) {
	var resp Response[Function]
	if err := c.get(functionsPath+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteFunction(id string) error {
	return c.delete(functionsPath+id, nil)
}

func (c *Client) InvokeFunction(id string, payload []byte, async bool) (*Invocation, error) {
	url := functionsPath + id + "/invoke"
	if async {
		url += "?async=true"
	}

	var resp Response[Invocation]
	// post helper expects a struct for body, but we have []byte.
	// We'll use resty directly or fix helper.
	// The post helper in client.go: req.SetBody(body). Resty handles []byte.
	if err := c.post(url, payload, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetFunctionLogs(id string) ([]*Invocation, error) {
	var resp Response[[]*Invocation]
	if err := c.get(functionsPath+id+"/logs", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
