// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// NotifyRepository manages the persistent state of notification topics and subscriptions.
type NotifyRepository interface {
	// CreateTopic saves a new notification topic.
	CreateTopic(ctx context.Context, topic *domain.Topic) error
	// GetTopicByID retrieves a topic by its unique UUID.
	GetTopicByID(ctx context.Context, id, userID uuid.UUID) (*domain.Topic, error)
	// GetTopicByName retrieves a topic by its friendly name for a specific user.
	GetTopicByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Topic, error)
	// ListTopics returns all topics owned by a user.
	ListTopics(ctx context.Context, userID uuid.UUID) ([]*domain.Topic, error)
	// DeleteTopic removes a topic and all its associated subscriptions.
	DeleteTopic(ctx context.Context, id uuid.UUID) error

	// CreateSubscription saves a new link between a topic and a delivery endpoint.
	CreateSubscription(ctx context.Context, sub *domain.Subscription) error
	// GetSubscriptionByID retrieves a subscription by its unique UUID.
	GetSubscriptionByID(ctx context.Context, id, userID uuid.UUID) (*domain.Subscription, error)
	// ListSubscriptions returns all active delivery points for a specific topic.
	ListSubscriptions(ctx context.Context, topicID uuid.UUID) ([]*domain.Subscription, error)
	// DeleteSubscription removes a message delivery endpoint.
	DeleteSubscription(ctx context.Context, id uuid.UUID) error

	// For message delivery

	// SaveMessage records a history of a message published to a topic.
	SaveMessage(ctx context.Context, msg *domain.NotifyMessage) error
}

// NotifyService provides business logic for the notification and messaging system (e.g., SNS-like).
type NotifyService interface {
	// CreateTopic establishes a new broadcast channel.
	CreateTopic(ctx context.Context, name string) (*domain.Topic, error)
	// ListTopics returns all broadcast channels for the current authorized user.
	ListTopics(ctx context.Context) ([]*domain.Topic, error)
	// DeleteTopic decommissioning a broadcast channel.
	DeleteTopic(ctx context.Context, id uuid.UUID) error

	// Subscribe links a notification channel to a target endpoint (Queue or Webhook).
	Subscribe(ctx context.Context, topicID uuid.UUID, protocol domain.SubscriptionProtocol, endpoint string) (*domain.Subscription, error)
	// ListSubscriptions returns all delivery targets for a specific channel.
	ListSubscriptions(ctx context.Context, topicID uuid.UUID) ([]*domain.Subscription, error)
	// Unsubscribe removes a delivery link from a notification channel.
	Unsubscribe(ctx context.Context, id uuid.UUID) error

	// Publish broadcasts a message to all subscribers of a specific topic.
	Publish(ctx context.Context, topicID uuid.UUID, body string) error
}
