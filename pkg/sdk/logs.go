package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// LogEntry represents a single log line in the CloudLogs service.
type LogEntry struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	TraceID      string    `json:"trace_id,omitempty"`
}

// LogSearchResponse represents the paginated response for log searches.
type LogSearchResponse struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
}

// LogQuery contains filters for log searching.
type LogQuery struct {
	ResourceID   string
	ResourceType string
	Level        string
	Search       string
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
	Offset       int
}

// SearchLogs searches and filters platform logs.
func (c *Client) SearchLogs(ctx context.Context, query LogQuery) (*LogSearchResponse, error) {
	params := url.Values{}
	if query.ResourceID != "" {
		params.Add("resource_id", query.ResourceID)
	}
	if query.ResourceType != "" {
		params.Add("resource_type", query.ResourceType)
	}
	if query.Level != "" {
		params.Add("level", query.Level)
	}
	if query.Search != "" {
		params.Add("search", query.Search)
	}
	if query.StartTime != nil {
		params.Add("start_time", query.StartTime.Format(time.RFC3339))
	}
	if query.EndTime != nil {
		params.Add("end_time", query.EndTime.Format(time.RFC3339))
	}
	if query.Limit > 0 {
		params.Add("limit", strconv.Itoa(query.Limit))
	}
	if query.Offset > 0 {
		params.Add("offset", strconv.Itoa(query.Offset))
	}

	path := "/logs"
	if params.Encode() != "" {
		path += "?" + params.Encode()
	}

	var res Response[LogSearchResponse]
	if err := c.getWithContext(ctx, path, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// GetLogsByResource retrieves logs for a specific resource ID.
func (c *Client) GetLogsByResource(ctx context.Context, resourceID string, limit int) (*LogSearchResponse, error) {
	path := fmt.Sprintf("/logs/%s", resourceID)
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}

	var res Response[LogSearchResponse]
	if err := c.getWithContext(ctx, path, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}
