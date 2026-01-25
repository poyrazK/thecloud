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

type fakeTaskQueue struct {
	messages []string
	errors   []error // To simulate dequeue errors
	index    int
}

func (f *fakeTaskQueue) Enqueue(ctx context.Context, queueName string, payload interface{}) error {
	return nil
}

func (f *fakeTaskQueue) Dequeue(ctx context.Context, queueName string) (string, error) {
	if f.index < len(f.errors) && f.errors[f.index] != nil {
		err := f.errors[f.index]
		f.index++
		return "", err
	}
	if f.index < len(f.messages) {
		msg := f.messages[f.index]
		f.index++
		return msg, nil
	}
	return "", nil
}

// failingComputeBackend forces Provision to fail
type failingComputeBackend struct {
	noop.NoopComputeBackend
}

func (f *failingComputeBackend) CreateInstance(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	return "", errors.New("provisioning failed")
}

func TestProvisionWorker_Run(t *testing.T) {
	tests := []struct {
		name           string
		message        interface{} // string or struct
		injectDequeErr bool
		failProvision  bool
		wantLog        string
	}{
		{
			name: "success",
			message: domain.ProvisionJob{
				InstanceID: uuid.New(),
				UserID:     uuid.New(),
			},
			wantLog: "successfully provisioned instance",
		},
		{
			name:    "deserialize_error",
			message: "{invalid-json}",
			wantLog: "failed to unmarshal provision job",
		},
		{
			name:          "provision_error",
			message:       domain.ProvisionJob{InstanceID: uuid.New(), UserID: uuid.New()},
			failProvision: true,
			wantLog:       "failed to provision instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msgBytes []byte
			switch v := tt.message.(type) {
			case string:
				msgBytes = []byte(v)
			default:
				msgBytes, _ = json.Marshal(v)
			}

			fq := &fakeTaskQueue{
				messages: []string{string(msgBytes)},
			}

			// Compute backend
			var compute ports.ComputeBackend = &noop.NoopComputeBackend{}
			if tt.failProvision {
				compute = &failingComputeBackend{}
			}

			instSvc := services.NewInstanceService(services.InstanceServiceParams{
				Repo:       &noop.NoopInstanceRepository{},
				VpcRepo:    &noop.NoopVpcRepository{},
				SubnetRepo: &noop.NoopSubnetRepository{},
				VolumeRepo: &noop.NoopVolumeRepository{},
				Compute:    compute,
				Network:    noop.NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))),
				EventSvc:   &noop.NoopEventService{},
				AuditSvc:   &noop.NoopAuditService{},
				TaskQueue:  nil,
				Logger:     slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
			})

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			worker := NewProvisionWorker(instSvc, fq, logger)

			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			wg.Add(1)

			go worker.Run(ctx, &wg)

			time.Sleep(50 * time.Millisecond)
			cancel()
			wg.Wait()

			assert.Contains(t, buf.String(), tt.wantLog)
		})
	}
}

func TestProvisionWorker_Run_DequeueError(t *testing.T) {
	// Test that worker continues on queue error
	fq := &fakeTaskQueue{
		messages: []string{},
		errors:   []error{errors.New("redis connection failed")},
	}

	instSvc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:       &noop.NoopInstanceRepository{},
		VpcRepo:    &noop.NoopVpcRepository{},
		SubnetRepo: &noop.NoopSubnetRepository{},
		VolumeRepo: &noop.NoopVolumeRepository{},
		Compute:    &noop.NoopComputeBackend{},
		Network:    noop.NewNoopNetworkAdapter(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))),
		EventSvc:   &noop.NoopEventService{},
		AuditSvc:   &noop.NoopAuditService{},
		TaskQueue:  nil,
		Logger:     slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	})

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	worker := NewProvisionWorker(instSvc, fq, logger)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go worker.Run(ctx, &wg)
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()
	// No specific log to check as it just continues, but we ensure no panic and coverage hits error path
}
