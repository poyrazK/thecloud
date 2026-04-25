package platform

import (
	"context"
	"log/slog"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ResilientDNSOpts configures the resilient DNS wrapper.
type ResilientDNSOpts struct {
	CallTimeout    time.Duration // Per-call timeout. Default: 10s.
	CBThreshold    int           // Failures to trip. Default: 5.
	CBResetTimeout time.Duration // Open→half-open wait. Default: 30s.
}

func (o ResilientDNSOpts) withDefaults() ResilientDNSOpts {
	if o.CallTimeout <= 0 {
		o.CallTimeout = 10 * time.Second
	}
	if o.CBThreshold <= 0 {
		o.CBThreshold = 5
	}
	if o.CBResetTimeout <= 0 {
		o.CBResetTimeout = 30 * time.Second
	}
	return o
}

// ResilientDNS wraps a DNSBackend with circuit breaker and per-call timeouts.
// DNS calls are lightweight so no bulkhead is applied (PowerDNS HTTP API is
// already serialized).
type ResilientDNS struct {
	inner  ports.DNSBackend
	cb     *CircuitBreaker
	logger *slog.Logger
	opts   ResilientDNSOpts
}

// NewResilientDNS decorates inner with resilience primitives.
func NewResilientDNS(inner ports.DNSBackend, logger *slog.Logger, opts ResilientDNSOpts) *ResilientDNS {
	opts = opts.withDefaults()
	name := "dns-powerdns"

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

	return &ResilientDNS{
		inner:  inner,
		cb:     cb,
		logger: logger.With("adapter", name),
		opts:   opts,
	}
}

func (r *ResilientDNS) callProtected(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, r.opts.CallTimeout)
		defer cancel()
		return fn(ctx2)
	})
}

// ---------- Zone Operations ----------

func (r *ResilientDNS) CreateZone(ctx context.Context, zoneName string, nameservers []string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.CreateZone(ctx, zoneName, nameservers)
	})
}

func (r *ResilientDNS) DeleteZone(ctx context.Context, zoneName string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteZone(ctx, zoneName)
	})
}

func (r *ResilientDNS) GetZone(ctx context.Context, zoneName string) (*ports.ZoneInfo, error) {
	var info *ports.ZoneInfo
	err := r.callProtected(ctx, func(ctx context.Context) error {
		var e error
		info, e = r.inner.GetZone(ctx, zoneName)
		return e
	})
	return info, err
}

// ---------- Record Operations ----------

func (r *ResilientDNS) AddRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.AddRecords(ctx, zoneName, records)
	})
}

func (r *ResilientDNS) UpdateRecords(ctx context.Context, zoneName string, records []ports.RecordSet) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.UpdateRecords(ctx, zoneName, records)
	})
}

func (r *ResilientDNS) DeleteRecords(ctx context.Context, zoneName string, name string, recordType string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteRecords(ctx, zoneName, name, recordType)
	})
}

func (r *ResilientDNS) ListRecords(ctx context.Context, zoneName string) ([]ports.RecordSet, error) {
	var records []ports.RecordSet
	err := r.callProtected(ctx, func(ctx context.Context) error {
		var e error
		records, e = r.inner.ListRecords(ctx, zoneName)
		return e
	})
	return records, err
}
