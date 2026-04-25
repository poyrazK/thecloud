package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestDurableQueue(t *testing.T) (*durableTaskQueue, *miniredis.Miniredis) {
	t.Helper()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	q := NewDurableTaskQueue(client, WithBlockTime(100*time.Millisecond), WithMaxRetries(3))
	return q, s
}

func TestDurableEnqueue(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	payload := map[string]string{"instance_id": "abc-123"}

	if err := q.Enqueue(ctx, "test_stream", payload); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Verify stream has one entry
	entries, err := s.Stream("test_stream")
	if err != nil {
		t.Fatalf("Stream read failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 stream entry, got %d", len(entries))
	}
}

func TestDurableEnsureGroupIdempotent(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	// Should not error even if stream doesn't exist yet (MkStream).
	if err := q.EnsureGroup(ctx, "test_stream", "workers"); err != nil {
		t.Fatalf("first EnsureGroup failed: %v", err)
	}
	// Calling again should be idempotent (BUSYGROUP).
	if err := q.EnsureGroup(ctx, "test_stream", "workers"); err != nil {
		t.Fatalf("second EnsureGroup failed: %v", err)
	}
}

func TestDurableReceiveAndAck(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	queue := "provision_queue"
	group := "workers"
	consumer := "worker-1"

	// Setup
	if err := q.EnsureGroup(ctx, queue, group); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	// Enqueue a job
	job := map[string]string{"instance_id": "inst-001"}
	if err := q.Enqueue(ctx, queue, job); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	// Receive it
	msg, err := q.Receive(ctx, queue, group, consumer)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if msg == nil {
		t.Fatal("expected message, got nil")
	}
	if msg.Queue != queue {
		t.Fatalf("expected queue %q, got %q", queue, msg.Queue)
	}

	// Verify payload round-trips
	var got map[string]string
	if err := json.Unmarshal([]byte(msg.Payload), &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got["instance_id"] != "inst-001" {
		t.Fatalf("expected instance_id inst-001, got %s", got["instance_id"])
	}

	// Ack it
	if err := q.Ack(ctx, queue, group, msg.ID); err != nil {
		t.Fatalf("Ack: %v", err)
	}

	// Receive again — should be empty
	msg2, err := q.Receive(ctx, queue, group, consumer)
	if err != nil {
		t.Fatalf("second Receive: %v", err)
	}
	if msg2 != nil {
		t.Fatalf("expected nil after ack, got %+v", msg2)
	}
}

func TestDurableReceiveEmptyReturnsNil(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	queue := "empty_stream"
	group := "workers"

	if err := q.EnsureGroup(ctx, queue, group); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	msg, err := q.Receive(ctx, queue, group, "worker-1")
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil message from empty stream, got %+v", msg)
	}
}

func TestDurableMultipleConsumersGetDifferentMessages(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	queue := "multi_consumer"
	group := "workers"

	if err := q.EnsureGroup(ctx, queue, group); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	// Enqueue two messages
	if err := q.Enqueue(ctx, queue, map[string]string{"id": "1"}); err != nil {
		t.Fatalf("Enqueue 1: %v", err)
	}
	if err := q.Enqueue(ctx, queue, map[string]string{"id": "2"}); err != nil {
		t.Fatalf("Enqueue 2: %v", err)
	}

	// Two consumers each get one
	msg1, err := q.Receive(ctx, queue, group, "worker-1")
	if err != nil || msg1 == nil {
		t.Fatalf("worker-1 Receive: msg=%v err=%v", msg1, err)
	}
	msg2, err := q.Receive(ctx, queue, group, "worker-2")
	if err != nil || msg2 == nil {
		t.Fatalf("worker-2 Receive: msg=%v err=%v", msg2, err)
	}

	if msg1.ID == msg2.ID {
		t.Fatalf("both consumers got the same message ID: %s", msg1.ID)
	}
}

func TestDurableNackKeepsMessagePending(t *testing.T) {
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	queue := "nack_test"
	group := "workers"
	consumer := "worker-1"

	if err := q.EnsureGroup(ctx, queue, group); err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}
	if err := q.Enqueue(ctx, queue, map[string]string{"id": "1"}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	msg, err := q.Receive(ctx, queue, group, consumer)
	if err != nil || msg == nil {
		t.Fatalf("Receive: msg=%v err=%v", msg, err)
	}

	// Nack (no-op in Redis Streams — message stays in PEL)
	if err := q.Nack(ctx, queue, group, msg.ID); err != nil {
		t.Fatalf("Nack: %v", err)
	}

	// The message should still be pending (not acked).
	// Verify via XPending.
	pending, err := q.client.XPending(ctx, queue, group).Result()
	if err != nil {
		t.Fatalf("XPending: %v", err)
	}
	if pending.Count != 1 {
		t.Fatalf("expected 1 pending message, got %d", pending.Count)
	}
}

func TestDurableDeadLetterOnDequeue(t *testing.T) {
	// This tests the legacy Dequeue path for backward compatibility.
	q, s := newTestDurableQueue(t)
	defer s.Close()

	ctx := context.Background()
	queue := "legacy_dequeue"

	if err := q.Enqueue(ctx, queue, map[string]string{"legacy": "true"}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	msg, err := q.Dequeue(ctx, queue)
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}
	if msg == "" {
		t.Fatal("expected non-empty legacy dequeue result")
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(msg), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["legacy"] != "true" {
		t.Fatalf("expected legacy=true, got %s", got["legacy"])
	}

	// Stream should be empty after legacy Dequeue (auto-deleted)
	entries, err := s.Stream(queue)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty stream after legacy dequeue, got %d entries", len(entries))
	}
}
