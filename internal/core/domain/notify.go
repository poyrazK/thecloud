// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Topic represents a named communication channel for broadcasting messages.
// It follows a publish-subscribe pattern (Pub/Sub).
type Topic struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	ARN       string    `json:"arn"` // Unique identifier (arn:thecloud:notify:{region}:{user}:topic/{name})
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionProtocol defines the supported delivery methods for notifications.
type SubscriptionProtocol string

const (
	// ProtocolQueue delivers messages to a managed message queue.
	ProtocolQueue SubscriptionProtocol = "queue"
	// ProtocolWebhook delivers messages via an HTTP POST request to a provided URL.
	ProtocolWebhook SubscriptionProtocol = "webhook"
)

// Subscription represents a link between a Topic and a delivery endpoint.
type Subscription struct {
	ID        uuid.UUID            `json:"id"`
	UserID    uuid.UUID            `json:"user_id"`
	TopicID   uuid.UUID            `json:"topic_id"`
	Protocol  SubscriptionProtocol `json:"protocol"`
	Endpoint  string               `json:"endpoint"` // Target address (e.g., Queue ARN or Webhook URL)
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// NotifyMessage represents a single piece of content published to a topic.
type NotifyMessage struct {
	ID        uuid.UUID `json:"id"`
	TopicID   uuid.UUID `json:"topic_id"`
	Body      string    `json:"body"` // The actual content of the notification
	CreatedAt time.Time `json:"created_at"`
}
