// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents an audit/event stream entry.
type Event struct {
	ID           uuid.UUID       `json:"id"`
	Action       string          `json:"action"`
	ResourceID   string          `json:"resource_id"`
	ResourceType string          `json:"resource_type"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (c *Client) ListEvents() ([]Event, error) {
	var res Response[[]Event]
	if err := c.get("/events?limit=50", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
