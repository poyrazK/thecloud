// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
	"io"
	"time"
)

// Object describes an object stored in a bucket.
type Object struct {
	ID          string    `json:"id"`
	ARN         string    `json:"arn"`
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	SizeBytes   int64     `json:"size_bytes"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Client) ListObjects(bucket string) ([]Object, error) {
	var res Response[[]Object]
	if err := c.get(fmt.Sprintf("/storage/%s", bucket), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) UploadObject(bucket, key string, body io.Reader) error {
	resp, err := c.resty.R().
		SetBody(body).
		Put(fmt.Sprintf("%s/storage/%s/%s", c.apiURL, bucket, key))

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) DownloadObject(bucket, key string) (io.ReadCloser, error) {
	resp, err := c.resty.R().
		SetDoNotParseResponse(true).
		Get(fmt.Sprintf("%s/storage/%s/%s", c.apiURL, bucket, key))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		resp.RawBody().Close()
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode())
	}
	return resp.RawBody(), nil
}

func (c *Client) DeleteObject(bucket, key string) error {
	return c.delete(fmt.Sprintf("/storage/%s/%s", bucket, key), nil)
}
