// Package redis implements Redis-based repositories and data structures.
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTaskQueue struct {
	client *redis.Client
}

func NewRedisTaskQueue(client *redis.Client) *redisTaskQueue {
	return &redisTaskQueue{client: client}
}

func (q *redisTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, queueName, data).Err()
}

func (q *redisTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	// BRPop blocks until a message is available
	res, err := q.client.BRPop(ctx, 5*time.Second, queueName).Result()
	if err != nil {
		return "", err
	}
	// res[0] is the key, res[1] is the value
	if len(res) < 2 {
		return "", nil
	}
	return res[1], nil
}
