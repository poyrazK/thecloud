// Package redis implements Redis-based repositories and data structures.
package redis

import (
	"context"
	stdlib_errors "errors"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTaskQueue struct {
	client *redis.Client
}

// NewRedisTaskQueue creates a Redis-backed task queue.
func NewRedisTaskQueue(client *redis.Client) *redisTaskQueue {
	return &redisTaskQueue{client: client}
}

func (q *redisTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// Log raw payload
	// fmt.Printf("Enqueueing to %s: %s\n", queueName, string(data))
	// We don't have a logger here easily, but we can rely on return being nil.
	return q.client.LPush(ctx, queueName, data).Err()
}

func (q *redisTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	// BRPop blocks until a message is available
	res, err := q.client.BRPop(ctx, 5*time.Second, queueName).Result()
	if err != nil {
		if stdlib_errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	// res[0] is the key, res[1] is the value
	if len(res) < 2 {
		return "", nil
	}
	return res[1], nil
}
