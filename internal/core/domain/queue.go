// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// QueueStatus represents the current state of a message queue.
type QueueStatus string

const (
	// QueueStatusActive indicates the queue is operational and accepting messages.
	QueueStatusActive QueueStatus = "ACTIVE"
	// QueueStatusDeleting indicates the queue is being purged and removed.
	QueueStatusDeleting QueueStatus = "DELETING"
)

// Queue represents a point-to-point asynchronous communication channel (MaaS).
type Queue struct {
	ID                uuid.UUID   `json:"id"`
	UserID            uuid.UUID   `json:"user_id"`
	Name              string      `json:"name"`
	ARN               string      `json:"arn"`                // Unique identifier (arn:thecloud:queue:{region}:{user}:{name})
	VisibilityTimeout int         `json:"visibility_timeout"` // Seconds a message remains hidden after retrieval
	RetentionDays     int         `json:"retention_days"`     // Days before non-deleted messages are purged
	MaxMessageSize    int         `json:"max_message_size"`   // Maximum payload size in bytes
	Status            QueueStatus `json:"status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// Message represents an individual data packet stored within a Queue.
type Message struct {
	ID            uuid.UUID `json:"id"`
	QueueID       uuid.UUID `json:"queue_id"`
	Body          string    `json:"body"`           // The message payload content
	ReceiptHandle string    `json:"receipt_handle"` // Unique identifier used to delete the message after processing
	VisibleAt     time.Time `json:"visible_at"`     // When the message becomes available for retrieval again
	ReceivedCount int       `json:"received_count"` // Number of times this message has been retrieved
	CreatedAt     time.Time `json:"created_at"`
}
