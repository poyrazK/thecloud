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
	VersionID   string    `json:"version_id"`
	IsLatest    bool      `json:"is_latest"`
	SizeBytes   int64     `json:"size_bytes"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// Bucket describes a storage bucket.
type Bucket struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	IsPublic          bool      `json:"is_public"`
	VersioningEnabled bool      `json:"versioning_enabled"`
	CreatedAt         time.Time `json:"created_at"`
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

type LifecycleRule struct {
	ID             string    `json:"id"`
	BucketName     string    `json:"bucket_name"`
	Prefix         string    `json:"prefix"`
	ExpirationDays int       `json:"expiration_days"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

func (c *Client) ListObjects(bucket string) ([]Object, error) {
	var res Response[[]Object]
	if err := c.get(fmt.Sprintf("/storage/%s", bucket), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) UploadObject(bucket, key string, body io.Reader) (*Object, error) {
	var res Response[Object]
	resp, err := c.resty.R().
		SetBody(body).
		SetResult(&res).
		Put(fmt.Sprintf("%s/storage/%s/%s", c.apiURL, bucket, key))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &res.Data, nil
}

func (c *Client) ListVersions(bucket, key string) ([]Object, error) {
	var res Response[[]Object]
	if err := c.get(fmt.Sprintf("/storage/versions/%s/%s", bucket, key), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) DownloadObject(bucket, key string, versionID ...string) (io.ReadCloser, error) {
	req := c.resty.R().SetDoNotParseResponse(true)
	if len(versionID) > 0 {
		req.SetQueryParam("versionId", versionID[0])
	}

	resp, err := req.Get(fmt.Sprintf("%s/storage/%s/%s", c.apiURL, bucket, key))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		resp.RawBody().Close()
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode())
	}
	return resp.RawBody(), nil
}

func (c *Client) DeleteObject(bucket, key string, versionID ...string) error {
	path := fmt.Sprintf("/storage/%s/%s", bucket, key)
	if len(versionID) > 0 {
		path += "?versionId=" + versionID[0]
	}
	return c.delete(path, nil)
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

func (c *Client) SetBucketVersioning(name string, enabled bool) error {
	req := struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: enabled,
	}
	return c.patch(fmt.Sprintf("/storage/buckets/%s/versioning", name), req, nil)
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

func (c *Client) CreateLifecycleRule(bucket, prefix string, expirationDays int, enabled bool) (*LifecycleRule, error) {
	req := struct {
		Prefix         string `json:"prefix"`
		ExpirationDays int    `json:"expiration_days"`
		Enabled        bool   `json:"enabled"`
	}{
		Prefix:         prefix,
		ExpirationDays: expirationDays,
		Enabled:        enabled,
	}
	var res Response[LifecycleRule]
	if err := c.post(fmt.Sprintf("/storage/buckets/%s/lifecycle", bucket), req, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func (c *Client) ListLifecycleRules(bucket string) ([]LifecycleRule, error) {
	var res Response[[]LifecycleRule]
	if err := c.get(fmt.Sprintf("/storage/buckets/%s/lifecycle", bucket), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) DeleteLifecycleRule(bucket, ruleID string) error {
	return c.delete(fmt.Sprintf("/storage/buckets/%s/lifecycle/%s", bucket, ruleID), nil)
}
