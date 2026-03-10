package platform

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ResilientLBOpts configures the resilient load balancer proxy wrapper.
type ResilientLBOpts struct {
	CallTimeout    time.Duration // Per-call timeout. Default: 30s.
	LongTimeout    time.Duration // Timeout for DeployProxy (container launch). Default: 2m.
	CBThreshold    int           // Failures to trip. Default: 5.
	CBResetTimeout time.Duration // Open→half-open wait. Default: 30s.
}

func (o ResilientLBOpts) withDefaults() ResilientLBOpts {
	if o.CallTimeout <= 0 {
		o.CallTimeout = 30 * time.Second
	}
	if o.LongTimeout <= 0 {
		o.LongTimeout = 2 * time.Minute
	}
	if o.CBThreshold <= 0 {
		o.CBThreshold = 5
	}
	if o.CBResetTimeout <= 0 {
		o.CBResetTimeout = 30 * time.Second
	}
	return o
}

// ResilientLB wraps an LBProxyAdapter with circuit breaker and per-call timeouts.
// LB proxy has only 3 methods so no bulkhead is needed — the compute bulkhead
// already limits the underlying container/VM creation.
type ResilientLB struct {
	inner  ports.LBProxyAdapter
	cb     *CircuitBreaker
	logger *slog.Logger
	opts   ResilientLBOpts
}

// NewResilientLB decorates inner with resilience primitives.
func NewResilientLB(inner ports.LBProxyAdapter, logger *slog.Logger, opts ResilientLBOpts) *ResilientLB {
	opts = opts.withDefaults()
	name := "lb-proxy"

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

	return &ResilientLB{
		inner:  inner,
		cb:     cb,
		logger: logger.With("adapter", name),
		opts:   opts,
	}
}

func (r *ResilientLB) DeployProxy(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) (string, error) {
	var addr string
	err := r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, r.opts.LongTimeout)
		defer cancel()
		var e error
		addr, e = r.inner.DeployProxy(ctx2, lb, targets)
		return e
	})
	return addr, err
}

func (r *ResilientLB) RemoveProxy(ctx context.Context, lbID uuid.UUID) error {
	return r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, r.opts.CallTimeout)
		defer cancel()
		return r.inner.RemoveProxy(ctx2, lbID)
	})
}

func (r *ResilientLB) UpdateProxyConfig(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) error {
	return r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, r.opts.CallTimeout)
		defer cancel()
		return r.inner.UpdateProxyConfig(ctx2, lb, targets)
	})
}
