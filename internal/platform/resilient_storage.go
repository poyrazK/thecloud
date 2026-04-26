package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ResilientStorageOpts configures the resilient storage wrapper.
type ResilientStorageOpts struct {
	CallTimeout     time.Duration // Per-call timeout. Default: 30s.
	LongCallTimeout time.Duration // Timeout for snapshot/restore. Default: 5m.
	CBThreshold     int           // Failures to trip. Default: 5.
	CBResetTimeout  time.Duration // Open→half-open wait. Default: 30s.
	BulkheadMaxConc int           // Max concurrent calls. Default: 10.
	BulkheadWait    time.Duration // Bulkhead slot wait. Default: 10s.
}

func (o ResilientStorageOpts) withDefaults() ResilientStorageOpts {
	if o.CallTimeout <= 0 {
		o.CallTimeout = 30 * time.Second
	}
	if o.LongCallTimeout <= 0 {
		o.LongCallTimeout = 5 * time.Minute
	}
	if o.CBThreshold <= 0 {
		o.CBThreshold = 5
	}
	if o.CBResetTimeout <= 0 {
		o.CBResetTimeout = 30 * time.Second
	}
	if o.BulkheadMaxConc <= 0 {
		o.BulkheadMaxConc = 10
	}
	if o.BulkheadWait <= 0 {
		o.BulkheadWait = 10 * time.Second
	}
	return o
}

// ResilientStorage wraps a StorageBackend with circuit breaker, bulkhead,
// and per-call timeouts.
type ResilientStorage struct {
	inner    ports.StorageBackend
	cb       *CircuitBreaker
	bulkhead *Bulkhead
	logger   *slog.Logger
	opts     ResilientStorageOpts
}

// NewResilientStorage decorates inner with resilience primitives.
func NewResilientStorage(inner ports.StorageBackend, logger *slog.Logger, opts ResilientStorageOpts) *ResilientStorage {
	opts = opts.withDefaults()
	name := fmt.Sprintf("storage-%s", inner.Type())

	cb := NewCircuitBreakerWithOpts(CircuitBreakerOpts{
		Name:            name,
		Threshold:       opts.CBThreshold,
		ResetTimeout:    opts.CBResetTimeout,
		SuccessRequired: 2,
		OnStateChange: func(n string, from, to State) {
			logger.Warn("circuit breaker state change",
				"breaker", n, "from", from.String(), "to", to.String())
		},
	})

	bh := NewBulkhead(BulkheadOpts{
		Name:        name,
		MaxConc:     opts.BulkheadMaxConc,
		WaitTimeout: opts.BulkheadWait,
	})

	return &ResilientStorage{
		inner:    inner,
		cb:       cb,
		bulkhead: bh,
		logger:   logger.With("adapter", name),
		opts:     opts,
	}
}

func (r *ResilientStorage) callProtected(ctx context.Context, timeout time.Duration, fn func(ctx context.Context) error) error {
	return r.bulkhead.Execute(ctx, func() error {
		return r.cb.Execute(func() error {
			ctx2, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return fn(ctx2)
		})
	})
}

func (r *ResilientStorage) CreateVolume(ctx context.Context, name string, sizeGB int) (string, error) {
	var path string
	err := r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		var e error
		path, e = r.inner.CreateVolume(ctx, name, sizeGB)
		return e
	})
	return path, err
}

func (r *ResilientStorage) DeleteVolume(ctx context.Context, name string) error {
	return r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		return r.inner.DeleteVolume(ctx, name)
	})
}

func (r *ResilientStorage) ResizeVolume(ctx context.Context, name string, newSizeGB int) error {
	return r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		return r.inner.ResizeVolume(ctx, name, newSizeGB)
	})
}

func (r *ResilientStorage) AttachVolume(ctx context.Context, volumeName, instanceID string) (string, error) {
	var devPath string
	err := r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		var e error
		devPath, e = r.inner.AttachVolume(ctx, volumeName, instanceID)
		return e
	})
	return devPath, err
}

func (r *ResilientStorage) DetachVolume(ctx context.Context, volumeName, instanceID string) error {
	return r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		return r.inner.DetachVolume(ctx, volumeName, instanceID)
	})
}

func (r *ResilientStorage) CreateSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	return r.callProtected(ctx, r.opts.LongCallTimeout, func(ctx context.Context) error {
		return r.inner.CreateSnapshot(ctx, volumeName, snapshotName)
	})
}

func (r *ResilientStorage) RestoreSnapshot(ctx context.Context, volumeName, snapshotName string) error {
	return r.callProtected(ctx, r.opts.LongCallTimeout, func(ctx context.Context) error {
		return r.inner.RestoreSnapshot(ctx, volumeName, snapshotName)
	})
}

func (r *ResilientStorage) DeleteSnapshot(ctx context.Context, snapshotName string) error {
	return r.callProtected(ctx, r.opts.CallTimeout, func(ctx context.Context) error {
		return r.inner.DeleteSnapshot(ctx, snapshotName)
	})
}

func (r *ResilientStorage) Ping(ctx context.Context) error {
	return r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return r.inner.Ping(ctx2)
	})
}

func (r *ResilientStorage) Type() string {
	return r.inner.Type()
}

// Unwrap returns the underlying StorageBackend.
func (r *ResilientStorage) Unwrap() ports.StorageBackend {
	return r.inner
}
