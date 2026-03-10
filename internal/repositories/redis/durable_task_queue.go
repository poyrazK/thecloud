// Package redis implements Redis-based repositories and data structures.
package redis

import (
	"context"
	"encoding/json"
	stdlib_errors "errors"
	"fmt"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

// durableTaskQueue implements ports.DurableTaskQueue using Redis Streams
// and consumer groups for at-least-once delivery semantics.
type durableTaskQueue struct {
	client     *redis.Client
	blockTime  time.Duration // how long Receive blocks waiting for new messages
	maxRetries int64         // max delivery attempts before a message is dead-lettered
	dlqSuffix  string        // suffix appended to queue name for the dead-letter stream
}

// DurableQueueOption configures a durableTaskQueue.
type DurableQueueOption func(*durableTaskQueue)

// WithBlockTime sets the Receive block duration (default 5s).
func WithBlockTime(d time.Duration) DurableQueueOption {
	return func(q *durableTaskQueue) { q.blockTime = d }
}

// WithMaxRetries sets the max delivery count before dead-lettering (default 5).
func WithMaxRetries(n int64) DurableQueueOption {
	return func(q *durableTaskQueue) { q.maxRetries = n }
}

// WithDLQSuffix sets the dead-letter queue suffix (default ":dlq").
func WithDLQSuffix(s string) DurableQueueOption {
	return func(q *durableTaskQueue) { q.dlqSuffix = s }
}

// NewDurableTaskQueue creates a Redis Streams–backed durable task queue.
func NewDurableTaskQueue(client *redis.Client, opts ...DurableQueueOption) *durableTaskQueue {
	q := &durableTaskQueue{
		client:     client,
		blockTime:  5 * time.Second,
		maxRetries: 5,
		dlqSuffix:  ":dlq",
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

// ---------- TaskQueue (backward-compatible) ----------

func (q *durableTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("durable enqueue marshal: %w", err)
	}
	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: queueName,
		Values: map[string]interface{}{"payload": string(data)},
	}).Err()
}

func (q *durableTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	// Legacy fallback: reads from the stream without consumer groups (XREAD).
	// New consumers should use Receive instead.
	res, err := q.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{queueName, "0-0"},
		Count:   1,
		Block:   q.blockTime,
	}).Result()
	if err != nil {
		if stdlib_errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	if len(res) == 0 || len(res[0].Messages) == 0 {
		return "", nil
	}
	msg := res[0].Messages[0]
	// Auto-delete since legacy callers don't ack.
	q.client.XDel(ctx, queueName, msg.ID)
	payload, _ := msg.Values["payload"].(string)
	return payload, nil
}

// ---------- DurableTaskQueue ----------

func (q *durableTaskQueue) EnsureGroup(ctx context.Context, queueName, groupName string) error {
	err := q.client.XGroupCreateMkStream(ctx, queueName, groupName, "0").Err()
	if err != nil {
		// "BUSYGROUP Consumer Group name already exists" is harmless at startup.
		if isGroupExistsErr(err) {
			return nil
		}
		return fmt.Errorf("ensure group %s/%s: %w", queueName, groupName, err)
	}
	return nil
}

func (q *durableTaskQueue) Receive(ctx context.Context, queueName, groupName, consumerName string) (*ports.DurableMessage, error) {
	res, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{queueName, ">"},
		Count:    1,
		Block:    q.blockTime,
	}).Result()
	if err != nil {
		if stdlib_errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("receive from %s/%s: %w", queueName, groupName, err)
	}
	if len(res) == 0 || len(res[0].Messages) == 0 {
		return nil, nil
	}

	xmsg := res[0].Messages[0]
	payload, _ := xmsg.Values["payload"].(string)
	return &ports.DurableMessage{
		ID:      xmsg.ID,
		Payload: payload,
		Queue:   queueName,
	}, nil
}

func (q *durableTaskQueue) Ack(ctx context.Context, queueName, groupName, messageID string) error {
	return q.client.XAck(ctx, queueName, groupName, messageID).Err()
}

func (q *durableTaskQueue) Nack(ctx context.Context, queueName, groupName, messageID string) error {
	// In Redis Streams, un-acknowledged messages remain in the PEL (Pending
	// Entries List) automatically. Nack is a no-op — the message will be
	// reclaimed by ReclaimStale after the idle timeout.
	//
	// Future enhancement: we could XCLAIM the message back to a retry consumer
	// immediately, but the idle-reclaim approach is simpler and sufficient.
	return nil
}

func (q *durableTaskQueue) ReclaimStale(ctx context.Context, queueName, groupName, consumerName string, minIdleMs int64, count int64) ([]ports.DurableMessage, error) {
	// XAUTOCLAIM atomically claims messages idle > minIdleMs and returns them.
	msgs, _, err := q.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   queueName,
		Group:    groupName,
		Consumer: consumerName,
		MinIdle:  time.Duration(minIdleMs) * time.Millisecond,
		Start:    "0-0",
		Count:    count,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("reclaim stale from %s/%s: %w", queueName, groupName, err)
	}

	out := make([]ports.DurableMessage, 0, len(msgs))
	for _, xmsg := range msgs {
		payload, _ := xmsg.Values["payload"].(string)

		// Dead-letter messages that exceeded max retries.
		if xmsg.DeliveredCount > 0 && xmsg.DeliveredCount > q.maxRetries {
			_ = q.deadLetter(ctx, queueName, groupName, xmsg)
			continue
		}

		out = append(out, ports.DurableMessage{
			ID:      xmsg.ID,
			Payload: payload,
			Queue:   queueName,
		})
	}
	return out, nil
}

// deadLetter moves a message to the dead-letter stream and acks the original.
func (q *durableTaskQueue) deadLetter(ctx context.Context, queueName, groupName string, msg redis.XMessage) error {
	dlq := queueName + q.dlqSuffix
	payload, _ := msg.Values["payload"].(string)
	pipe := q.client.Pipeline()
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: dlq,
		Values: map[string]interface{}{
			"payload":     payload,
			"original_id": msg.ID,
			"queue":       queueName,
		},
	})
	pipe.XAck(ctx, queueName, groupName, msg.ID)
	pipe.XDel(ctx, queueName, msg.ID)
	_, err := pipe.Exec(ctx)
	return err
}

// isGroupExistsErr returns true when the error indicates the consumer group
// already exists (Redis returns BUSYGROUP).
func isGroupExistsErr(err error) bool {
	if err == nil {
		return false
	}
	return containsBusyGroup(err.Error())
}

func containsBusyGroup(s string) bool {
	return len(s) >= 9 && (s[:9] == "BUSYGROUP" || containsSubstring(s, "BUSYGROUP"))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
