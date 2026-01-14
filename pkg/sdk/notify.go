// Package sdk provides the official Go SDK for the platform.
package sdk

import (
	"fmt"
)

// Topic describes a notification topic.
type Topic struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	ARN       string `json:"arn"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Subscription describes a topic subscription.
type Subscription struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TopicID   string `json:"topic_id"`
	Protocol  string `json:"protocol"`
	Endpoint  string `json:"endpoint"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (c *Client) CreateTopic(name string) (*Topic, error) {
	req := struct {
		Name string `json:"name"`
	}{Name: name}

	var resp Response[Topic]
	err := c.post("/notify/topics", req, &resp)
	return &resp.Data, err
}

func (c *Client) ListTopics() ([]Topic, error) {
	var resp Response[[]Topic]
	err := c.get("/notify/topics", &resp)
	return resp.Data, err
}

func (c *Client) DeleteTopic(id string) error {
	return c.delete(fmt.Sprintf("/notify/topics/%s", id), nil)
}

func (c *Client) Subscribe(topicID, protocol, endpoint string) (*Subscription, error) {
	req := struct {
		Protocol string `json:"protocol"`
		Endpoint string `json:"endpoint"`
	}{Protocol: protocol, Endpoint: endpoint}

	var resp Response[Subscription]
	err := c.post(fmt.Sprintf("/notify/topics/%s/subscriptions", topicID), req, &resp)
	return &resp.Data, err
}

func (c *Client) ListSubscriptions(topicID string) ([]Subscription, error) {
	var resp Response[[]Subscription]
	err := c.get(fmt.Sprintf("/notify/topics/%s/subscriptions", topicID), &resp)
	return resp.Data, err
}

func (c *Client) Unsubscribe(id string) error {
	return c.delete(fmt.Sprintf("/notify/subscriptions/%s", id), nil)
}

func (c *Client) Publish(topicID, message string) error {
	req := struct {
		Message string `json:"message"`
	}{Message: message}

	return c.post(fmt.Sprintf("/notify/topics/%s/publish", topicID), req, nil)
}
