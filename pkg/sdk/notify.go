package sdk

import (
	"fmt"
)

type Topic struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	ARN       string `json:"arn"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

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

	var topic Topic
	err := c.post("/notify/topics", req, &topic)
	return &topic, err
}

func (c *Client) ListTopics() ([]Topic, error) {
	var topics []Topic
	err := c.get("/notify/topics", &topics)
	return topics, err
}

func (c *Client) DeleteTopic(id string) error {
	return c.delete(fmt.Sprintf("/notify/topics/%s", id), nil)
}

func (c *Client) Subscribe(topicID, protocol, endpoint string) (*Subscription, error) {
	req := struct {
		Protocol string `json:"protocol"`
		Endpoint string `json:"endpoint"`
	}{Protocol: protocol, Endpoint: endpoint}

	var sub Subscription
	err := c.post(fmt.Sprintf("/notify/topics/%s/subscriptions", topicID), req, &sub)
	return &sub, err
}

func (c *Client) ListSubscriptions(topicID string) ([]Subscription, error) {
	var subs []Subscription
	err := c.get(fmt.Sprintf("/notify/topics/%s/subscriptions", topicID), &subs)
	return subs, err
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
