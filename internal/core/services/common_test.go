package services_test

import (
	"context"
)

type TaskQueueStub struct{}

func (q *TaskQueueStub) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	return nil
}

func (q *TaskQueueStub) Dequeue(ctx context.Context, queueName string) (string, error) {
	return "", nil
}
