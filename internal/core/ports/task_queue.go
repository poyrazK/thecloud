// Package ports defines service and repository interfaces.
package ports

import (
	"context"
)

// TaskQueue defines a simple producer-consumer interface for background work distribution.
type TaskQueue interface {
	// Enqueue adds a serializable payload to the specified background processing queue.
	Enqueue(ctx context.Context, queueName string, payload interface{}) error
	// Dequeue pulls the next available raw message string from the background processing queue.
	Dequeue(ctx context.Context, queueName string) (string, error)
}
