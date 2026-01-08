// Package noop provides a no-op implementation of the NetworkBackend interface.
// This is used as a fallback when OVS is not available.
package noop

import (
	"context"
	"log/slog"

	"github.com/poyrazk/thecloud/internal/core/ports"
)

// NoopNetworkAdapter is a no-op implementation of NetworkBackend
type NoopNetworkAdapter struct {
	logger *slog.Logger
}

// NewNoopNetworkAdapter creates a new no-op network adapter
func NewNoopNetworkAdapter(logger *slog.Logger) *NoopNetworkAdapter {
	return &NoopNetworkAdapter{logger: logger}
}

func (n *NoopNetworkAdapter) CreateBridge(ctx context.Context, name string, vxlanID int) error {
	n.logger.Warn("noop network adapter: CreateBridge called but not implemented", "name", name)
	return nil
}

func (n *NoopNetworkAdapter) DeleteBridge(ctx context.Context, name string) error {
	n.logger.Warn("noop network adapter: DeleteBridge called but not implemented", "name", name)
	return nil
}

func (n *NoopNetworkAdapter) ListBridges(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (n *NoopNetworkAdapter) AddPort(ctx context.Context, bridge, portName string) error {
	n.logger.Warn("noop network adapter: AddPort called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) DeletePort(ctx context.Context, bridge, portName string) error {
	n.logger.Warn("noop network adapter: DeletePort called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) CreateVXLANTunnel(ctx context.Context, bridge string, vni int, remoteIP string) error {
	n.logger.Warn("noop network adapter: CreateVXLANTunnel called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) DeleteVXLANTunnel(ctx context.Context, bridge string, remoteIP string) error {
	n.logger.Warn("noop network adapter: DeleteVXLANTunnel called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) AddFlowRule(ctx context.Context, bridge string, rule ports.FlowRule) error {
	n.logger.Warn("noop network adapter: AddFlowRule called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) DeleteFlowRule(ctx context.Context, bridge string, match string) error {
	n.logger.Warn("noop network adapter: DeleteFlowRule called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) ListFlowRules(ctx context.Context, bridge string) ([]ports.FlowRule, error) {
	return []ports.FlowRule{}, nil
}

func (n *NoopNetworkAdapter) CreateVethPair(ctx context.Context, hostEnd, containerEnd string) error {
	n.logger.Warn("noop network adapter: CreateVethPair called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error {
	n.logger.Warn("noop network adapter: AttachVethToBridge called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) DeleteVethPair(ctx context.Context, hostEnd string) error {
	n.logger.Warn("noop network adapter: DeleteVethPair called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) SetVethIP(ctx context.Context, vethEnd, ip, cidr string) error {
	n.logger.Warn("noop network adapter: SetVethIP called but not implemented")
	return nil
}

func (n *NoopNetworkAdapter) Ping(ctx context.Context) error {
	return nil
}

func (n *NoopNetworkAdapter) Type() string {
	return "noop"
}
