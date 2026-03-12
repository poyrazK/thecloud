package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/stretchr/testify/assert"
)

// fakeDurableQueue implements ports.DurableTaskQueue for testing.
type fakeDurableQueue struct {
	messages []*ports.DurableMessage
	errors   []error
	index    int
	acked    []string
	nacked   []string
}

func (f *fakeDurableQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	return nil
}

func (f *fakeDurableQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	return "", nil
}

func (f *fakeDurableQueue) EnsureGroup(ctx context.Context, queueName, groupName string) error {
	return nil
}

func (f *fakeDurableQueue) Receive(ctx context.Context, queueName, groupName, consumerName string) (*ports.DurableMessage, error) {
	if f.index < len(f.errors) && f.errors[f.index] != nil {
		err := f.errors[f.index]
		f.index++
		return nil, err
	}
	if f.index < len(f.messages) {
		msg := f.messages[f.index]
		f.index++
		return msg, nil
	}
	return nil, nil
}

func (f *fakeDurableQueue) Ack(ctx context.Context, queueName, groupName, messageID string) error {
	f.acked = append(f.acked, messageID)
	return nil
}

func (f *fakeDurableQueue) Nack(ctx context.Context, queueName, groupName, messageID string) error {
	f.nacked = append(f.nacked, messageID)
	return nil
}

func (f *fakeDurableQueue) ReclaimStale(ctx context.Context, queueName, groupName, consumerName string, minIdleMs int64, count int64) ([]ports.DurableMessage, error) {
	return nil, nil
}

// failingComputeBackend forces Provision to fail
type failingComputeBackend struct {
	noop.NoopComputeBackend
}

func (f *failingComputeBackend) LaunchInstanceWithOptions(ctx context.Context, opts ports.CreateInstanceOptions) (string, []string, error) {
	return "", nil, errors.New("provisioning failed")
}

func TestProvisionWorkerRun(t *testing.T) {
	tests := []struct {
		name          string
		payload       interface{}
		poisonJSON    bool
		failProvision bool
		wantLog       string
		wantAcked     bool
		wantNacked    bool
	}{
		{
			name: "success",
			payload: domain.ProvisionJob{
				InstanceID: uuid.New(),
				UserID:     uuid.New(),
			},
			wantLog:   "successfully provisioned instance",
			wantAcked: true,
		},
		{
			name:       "deserialize_error",
			poisonJSON: true,
			wantLog:    "failed to unmarshal provision job",
			wantAcked:  true, // poison messages are acked to unblock the queue
		},
		{
			name: "provision_error",
			payload: domain.ProvisionJob{
				InstanceID: uuid.New(),
				UserID:     uuid.New(),
			},
			failProvision: true,
			wantLog:       "failed to provision instance",
			wantNacked:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payloadStr string
			if tt.poisonJSON {
				payloadStr = "{invalid-json}"
			} else {
				data, _ := json.Marshal(tt.payload)
				payloadStr = string(data)
			}

			fq := &fakeDurableQueue{
				messages: []*ports.DurableMessage{
					{ID: "1-0", Payload: payloadStr, Queue: provisionQueue},
				},
			}

			var compute ports.ComputeBackend = &noop.NoopComputeBackend{}
			if tt.failProvision {
				compute = &failingComputeBackend{}
			}

			instSvc := services.NewInstanceService(services.InstanceServiceParams{
				Repo:             &noop.NoopInstanceRepository{},
				VpcRepo:          &noop.NoopVpcRepository{},
				SubnetRepo:       &noop.NoopSubnetRepository{},
				VolumeRepo:       &noop.NoopVolumeRepository{},
				InstanceTypeRepo: &noop.NoopInstanceTypeRepository{},
				Compute:          compute,
				Network:          noop.NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))),
				EventSvc:         &noop.NoopEventService{},
				AuditSvc:         &noop.NoopAuditService{},
				TaskQueue:        nil,
				Logger:           slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			})

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			worker := NewProvisionWorker(instSvc, fq, nil, logger)

			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			wg.Add(1)

			go worker.Run(ctx, &wg)

			// Give worker time to process
			time.Sleep(200 * time.Millisecond)
			cancel()
			wg.Wait()

			assert.Contains(t, buf.String(), tt.wantLog)
			if tt.wantAcked {
				assert.NotEmpty(t, fq.acked, "expected message to be acked")
				assert.Empty(t, fq.nacked, "did not expect message to be nacked when acked")
			} else if tt.wantNacked {
				assert.NotEmpty(t, fq.nacked, "expected message to be nacked")
				assert.Empty(t, fq.acked, "did not expect message to be acked when nacked")
			} else {
				assert.Empty(t, fq.acked, "expected no ack")
				assert.Empty(t, fq.nacked, "expected no nack")
			}
		})
	}
}

func TestProvisionWorkerRunReceiveError(t *testing.T) {
	fq := &fakeDurableQueue{
		errors: []error{errors.New("redis connection failed")},
	}

	instSvc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             &noop.NoopInstanceRepository{},
		VpcRepo:          &noop.NoopVpcRepository{},
		SubnetRepo:       &noop.NoopSubnetRepository{},
		VolumeRepo:       &noop.NoopVolumeRepository{},
		InstanceTypeRepo: &noop.NoopInstanceTypeRepository{},
		Compute:          &noop.NoopComputeBackend{},
		Network:          noop.NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))),
		EventSvc:         &noop.NoopEventService{},
		AuditSvc:         &noop.NoopAuditService{},
		TaskQueue:        nil,
		Logger:           slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	})

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	worker := NewProvisionWorker(instSvc, fq, nil, logger)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	assert.Contains(t, buf.String(), "failed to receive provision job")
}
