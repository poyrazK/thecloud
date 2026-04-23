package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ResilientNetworkOpts configures the resilient network wrapper.
type ResilientNetworkOpts struct {
	CallTimeout     time.Duration // Per-call timeout. Default: 30s.
	CBThreshold     int           // Failures to trip. Default: 5.
	CBResetTimeout  time.Duration // Open→half-open wait. Default: 30s.
	BulkheadMaxConc int           // Max concurrent calls. Default: 15.
	BulkheadWait    time.Duration // Bulkhead slot wait. Default: 10s.
}

func (o ResilientNetworkOpts) withDefaults() ResilientNetworkOpts {
	if o.CallTimeout <= 0 {
		o.CallTimeout = 30 * time.Second
	}
	if o.CBThreshold <= 0 {
		o.CBThreshold = 5
	}
	if o.CBResetTimeout <= 0 {
		o.CBResetTimeout = 30 * time.Second
	}
	if o.BulkheadMaxConc <= 0 {
		o.BulkheadMaxConc = 15
	}
	if o.BulkheadWait <= 0 {
		o.BulkheadWait = 10 * time.Second
	}
	return o
}

// ResilientNetwork wraps a NetworkBackend with circuit breaker, bulkhead,
// and per-call timeouts. It implements the ports.NetworkBackend interface.
type ResilientNetwork struct {
	inner    ports.NetworkBackend
	cb       *CircuitBreaker
	bulkhead *Bulkhead
	logger   *slog.Logger
	opts     ResilientNetworkOpts
}

// NewResilientNetwork decorates inner with resilience primitives.
func NewResilientNetwork(inner ports.NetworkBackend, logger *slog.Logger, opts ResilientNetworkOpts) *ResilientNetwork {
	opts = opts.withDefaults()
	name := fmt.Sprintf("network-%s", inner.Type())

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

	return &ResilientNetwork{
		inner:    inner,
		cb:       cb,
		bulkhead: bh,
		logger:   logger.With("adapter", name),
		opts:     opts,
	}
}

// callProtected runs fn through bulkhead → circuit breaker → timeout.
func (r *ResilientNetwork) callProtected(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.bulkhead.Execute(ctx, func() error {
		return r.cb.Execute(func() error {
			ctx2, cancel := context.WithTimeout(ctx, r.opts.CallTimeout)
			defer cancel()
			return fn(ctx2)
		})
	})
}

// ---------- Bridge Management ----------

func (r *ResilientNetwork) CreateBridge(ctx context.Context, name string, vxlanID int) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.CreateBridge(ctx, name, vxlanID)
	})
}

func (r *ResilientNetwork) DeleteBridge(ctx context.Context, name string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteBridge(ctx, name)
	})
}

func (r *ResilientNetwork) ListBridges(ctx context.Context) ([]string, error) {
	var bridges []string
	err := r.callProtected(ctx, func(ctx context.Context) error {
		var e error
		bridges, e = r.inner.ListBridges(ctx)
		return e
	})
	return bridges, err
}

// ---------- Port Management ----------

func (r *ResilientNetwork) AddPort(ctx context.Context, bridge, portName string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.AddPort(ctx, bridge, portName)
	})
}

func (r *ResilientNetwork) DeletePort(ctx context.Context, bridge, portName string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeletePort(ctx, bridge, portName)
	})
}

// ---------- VXLAN Tunnels ----------

func (r *ResilientNetwork) CreateVXLANTunnel(ctx context.Context, bridge string, vni int, remoteIP string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.CreateVXLANTunnel(ctx, bridge, vni, remoteIP)
	})
}

func (r *ResilientNetwork) DeleteVXLANTunnel(ctx context.Context, bridge string, remoteIP string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteVXLANTunnel(ctx, bridge, remoteIP)
	})
}

// ---------- Security Groups (Flow Rules) ----------

func (r *ResilientNetwork) AddFlowRule(ctx context.Context, bridge string, rule ports.FlowRule) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.AddFlowRule(ctx, bridge, rule)
	})
}

func (r *ResilientNetwork) DeleteFlowRule(ctx context.Context, bridge string, match string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteFlowRule(ctx, bridge, match)
	})
}

func (r *ResilientNetwork) ListFlowRules(ctx context.Context, bridge string) ([]ports.FlowRule, error) {
	var rules []ports.FlowRule
	err := r.callProtected(ctx, func(ctx context.Context) error {
		var e error
		rules, e = r.inner.ListFlowRules(ctx, bridge)
		return e
	})
	return rules, err
}

// ---------- Veth Pair Management ----------

func (r *ResilientNetwork) CreateVethPair(ctx context.Context, hostEnd, containerEnd string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.CreateVethPair(ctx, hostEnd, containerEnd)
	})
}

func (r *ResilientNetwork) AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.AttachVethToBridge(ctx, bridge, vethEnd)
	})
}

func (r *ResilientNetwork) DeleteVethPair(ctx context.Context, hostEnd string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.DeleteVethPair(ctx, hostEnd)
	})
}

func (r *ResilientNetwork) SetVethIP(ctx context.Context, vethEnd, ip, cidr string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.SetVethIP(ctx, vethEnd, ip, cidr)
	})
}

// ---------- Health ----------

func (r *ResilientNetwork) SetupNATForSubnet(ctx context.Context, bridge, natVethEnd, subnetCIDR, egressIP string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.SetupNATForSubnet(ctx, bridge, natVethEnd, subnetCIDR, egressIP)
	})
}

func (r *ResilientNetwork) RemoveNATForSubnet(ctx context.Context, bridge, natVethEnd, subnetCIDR, egressIP string) error {
	return r.callProtected(ctx, func(ctx context.Context) error {
		return r.inner.RemoveNATForSubnet(ctx, bridge, natVethEnd, subnetCIDR, egressIP)
	})
}

func (r *ResilientNetwork) Ping(ctx context.Context) error {
	return r.cb.Execute(func() error {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return r.inner.Ping(ctx2)
	})
}

func (r *ResilientNetwork) Type() string {
	return r.inner.Type()
}

// Unwrap returns the underlying NetworkBackend.
func (r *ResilientNetwork) Unwrap() ports.NetworkBackend {
	return r.inner
}
