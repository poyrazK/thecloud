// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// CreateQueueOptions encapsulates optional parameters for provisioning a new message queue.
type CreateQueueOptions struct {
	VisibilityTimeout *int // Seconds a message remains hidden after retrieval (overrides system default)
	RetentionDays     *int // Days before non-deleted messages are purged
	MaxMessageSize    *int // Maximum payload size in bytes
}

// QueueRepository handles the persistence of queue metadata and the low-level processing of messages.
type QueueRepository interface {
	// Create saves a new message queue record.
	Create(ctx context.Context, queue *domain.Queue) error
	// GetByID retrieves a queue by its unique UUID.
	GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Queue, error)
	// GetByName retrieves a queue by name for a specific user.
	GetByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Queue, error)
	// List returns all queues owned by a user.
	List(ctx context.Context, userID uuid.UUID) ([]*domain.Queue, error)
	// Delete removes a queue and its associated messages.
	Delete(ctx context.Context, id uuid.UUID) error

	// Messages

	// SendMessage inserts a new message into the queue.
	SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error)
	// ReceiveMessages retrieves a set of available messages from the queue.
	ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages, visibilityTimeout int) ([]*domain.Message, error)
	// DeleteMessage removes a message from the queue after successful processing via its receipt handle.
	DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error
	// PurgeMessages deletes every message currently in the queue without removing the queue itself.
	PurgeMessages(ctx context.Context, queueID uuid.UUID) (int64, error)
}

// QueueService provides business logic for point-to-point asynchronous messaging (e.g., SQS-like).
type QueueService interface {
	// CreateQueue provisions a new message delivery channel.
	CreateQueue(ctx context.Context, name string, opts *CreateQueueOptions) (*domain.Queue, error)
	// GetQueue retrieves details for a specific message queue.
	GetQueue(ctx context.Context, id uuid.UUID) (*domain.Queue, error)
	// ListQueues returns all message queues accessible to the authorized user.
	ListQueues(ctx context.Context) ([]*domain.Queue, error)
	// DeleteQueue decommissioning a message delivery channel.
	DeleteQueue(ctx context.Context, id uuid.UUID) error

	// Messages

	// SendMessage publishes a payload to the specified queue.
	SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error)
	// ReceiveMessages consumes messages from the queue for processing.
	ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages int) ([]*domain.Message, error)
	// DeleteMessage confirms successful processing and removes the message from the queue.
	DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error
	// PurgeQueue removes all existing messages from a queue.
	PurgeQueue(ctx context.Context, queueID uuid.UUID) error
}
