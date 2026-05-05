// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

// Client is the API client for the platform.
type Client struct {
	resty  *resty.Client
	apiURL string
	tenant string
}

const (
	errRequestFailed = "request failed: %w"
	errAPIError      = "api error: %s"
)

// ListResponse wraps paginated list responses.
type ListResponse[T any] struct {
	Data       []T   `json:"data"`
	TotalCount int   `json:"total_count,omitempty"`
	HasMore    bool  `json:"has_more,omitempty"`
}

// NewClient constructs a Client with the provided API URL and key.
func NewClient(apiURL, apiKey string) *Client {
	client := resty.New()
	client.SetHeader("X-API-Key", apiKey)
	client.SetRetryCount(3)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return false
		}
		statusCode := r.StatusCode()
		return statusCode >= 500 || statusCode == 429
	})
	return &Client{
		resty:  client,
		apiURL: apiURL,
	}
}

// EnableDebug enables debug mode for the underlying HTTP client.
func (c *Client) EnableDebug() {
	c.resty.SetDebug(true)
}

// SetTenant sets the X-Tenant-ID header for subsequent requests.
func (c *Client) SetTenant(id string) {
	c.tenant = id
	c.resty.SetHeader("X-Tenant-ID", id)
}

// Response wraps API responses returned by the platform.
type Response[T any] struct {
	Data  T              `json:"data"`
	Error *ErrorResponse `json:"error,omitempty"`
}

// ErrorResponse represents an API error payload.
type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// get performs a GET request against the API.
func (c *Client) get(path string, result interface{}) error {
	return c.getContext(context.Background(), path, result)
}

func (c *Client) getContext(ctx context.Context, path string, result interface{}) error {
	resp, err := c.resty.R().
		SetContext(ctx).
		SetResult(result).
		Get(c.apiURL + path)

	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}

func (c *Client) getWithContext(ctx context.Context, path string, result interface{}) error {
	return c.getContext(ctx, path, result)
}

// getWithPagination performs a GET request with optional pagination parameters.
func (c *Client) getWithPagination(path string, result interface{}, limit, offset int) error {
	return c.getContextWithPagination(context.Background(), path, result, limit, offset)
}

func (c *Client) getContextWithPagination(ctx context.Context, path string, result interface{}, limit, offset int) error {
	req := c.resty.R().SetContext(ctx).SetResult(result)
	if limit > 0 {
		req.SetQueryParam("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		req.SetQueryParam("offset", strconv.Itoa(offset))
	}

	resp, err := req.Get(c.apiURL + path)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	return c.postContext(context.Background(), path, body, result)
}

func (c *Client) postContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	req := c.resty.R().SetContext(ctx)
	if body != nil {
		req.SetBody(body)
	}
	if result != nil {
		req.SetResult(result)
	}

	resp, err := req.Post(c.apiURL + path)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}

func (c *Client) postWithContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.postContext(ctx, path, body, result)
}

func (c *Client) delete(path string, result interface{}) error {
	return c.deleteWithContext(context.Background(), path, result)
}

func (c *Client) deleteWithContext(ctx context.Context, path string, result interface{}) error {
	req := c.resty.R().SetContext(ctx)
	if result != nil {
		req.SetResult(result)
	}

	resp, err := req.Delete(c.apiURL + path)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}

func (c *Client) put(path string, body interface{}, result interface{}) error {
	return c.putWithContext(context.Background(), path, body, result)
}

func (c *Client) putWithContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	req := c.resty.R().SetContext(ctx)
	if body != nil {
		req.SetBody(body)
	}
	if result != nil {
		req.SetResult(result)
	}

	resp, err := req.Put(c.apiURL + path)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}

func (c *Client) patch(path string, body interface{}, result interface{}) error {
	return c.patchWithContext(context.Background(), path, body, result)
}

func (c *Client) patchWithContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	req := c.resty.R().SetContext(ctx)
	if body != nil {
		req.SetBody(body)
	}
	if result != nil {
		req.SetResult(result)
	}

	resp, err := req.Patch(c.apiURL + path)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}

	if resp.IsError() {
		return fmt.Errorf(errAPIError, resp.String())
	}

	return nil
}
