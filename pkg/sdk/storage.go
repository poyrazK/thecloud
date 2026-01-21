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

// Bucket describes a storage bucket.
type Bucket struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

type StorageNode struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

type StorageCluster struct {
	Nodes []StorageNode `json:"nodes"`
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

func (c *Client) CreateBucket(name string, isPublic bool) (*Bucket, error) {
	req := struct {
		Name     string `json:"name"`
		IsPublic bool   `json:"is_public"`
	}{
		Name:     name,
		IsPublic: isPublic,
	}
	var res Response[Bucket]
	if err := c.post("/storage/buckets", req, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func (c *Client) ListBuckets() ([]Bucket, error) {
	var res Response[[]Bucket]
	if err := c.get("/storage/buckets", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) DeleteBucket(name string) error {
	return c.delete(fmt.Sprintf("/storage/buckets/%s", name), nil)
}

func (c *Client) GetStorageClusterStatus() (*StorageCluster, error) {
	var res Response[StorageCluster]
	if err := c.get("/storage/cluster/status", &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

type PresignedURL struct {
	URL       string    `json:"url"`
	Method    string    `json:"method"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) GeneratePresignedURL(bucket, key, method string, expirySeconds int) (*PresignedURL, error) {
	req := struct {
		Method    string `json:"method"`
		ExpirySec int    `json:"expiry_seconds"`
	}{
		Method:    method,
		ExpirySec: expirySeconds,
	}

	var res Response[PresignedURL]
	if err := c.post(fmt.Sprintf("/storage/presign/%s/%s", bucket, key), req, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}
