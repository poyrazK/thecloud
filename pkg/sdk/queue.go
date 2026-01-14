// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
	"time"
)

// Queue describes a message queue.
type Queue struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ARN               string    `json:"arn"`
	VisibilityTimeout int       `json:"visibility_timeout"`
	RetentionDays     int       `json:"retention_days"`
	MaxMessageSize    int       `json:"max_message_size"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Message represents a queue message.
type Message struct {
	ID            string    `json:"id"`
	QueueID       string    `json:"queue_id"`
	Body          string    `json:"body"`
	ReceiptHandle string    `json:"receipt_handle,omitempty"`
	VisibleAt     time.Time `json:"visible_at"`
	ReceivedCount int       `json:"received_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func (c *Client) CreateQueue(name string, visibilityTimeout, retentionDays, maxMessageSize *int) (*Queue, error) {
	body := map[string]interface{}{
		"name": name,
	}
	if visibilityTimeout != nil {
		body["visibility_timeout"] = *visibilityTimeout
	}
	if retentionDays != nil {
		body["retention_days"] = *retentionDays
	}
	if maxMessageSize != nil {
		body["max_message_size"] = *maxMessageSize
	}

	var res Response[Queue]
	if err := c.post("/queues", body, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func (c *Client) ListQueues() ([]Queue, error) {
	var res Response[[]Queue]
	if err := c.get("/queues", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) GetQueue(id string) (*Queue, error) {
	var res Response[Queue]
	if err := c.get(fmt.Sprintf("/queues/%s", id), &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func (c *Client) DeleteQueue(id string) error {
	return c.delete(fmt.Sprintf("/queues/%s", id), nil)
}

func (c *Client) SendMessage(queueID string, body string) (*Message, error) {
	req := map[string]string{
		"body": body,
	}
	var res Response[Message]
	if err := c.post(fmt.Sprintf("/queues/%s/messages", queueID), req, &res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}

func (c *Client) ReceiveMessages(queueID string, maxMessages int) ([]Message, error) {
	var res Response[[]Message]
	url := fmt.Sprintf("/queues/%s/messages?max_messages=%d", queueID, maxMessages)
	if err := c.get(url, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) DeleteMessage(queueID string, receiptHandle string) error {
	return c.delete(fmt.Sprintf("/queues/%s/messages/%s", queueID, receiptHandle), nil)
}

func (c *Client) PurgeQueue(queueID string) error {
	return c.post(fmt.Sprintf("/queues/%s/purge", queueID), nil, nil)
}
