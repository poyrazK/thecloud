// Package ports defines service and repository interfaces.
package ports

import (
	"context"
)

// FlowRule represents an OpenFlow entry for Open vSwitch (OVS) to control packet forwarding.
type FlowRule struct {
	ID       string // Unique identifier for the rule
	Priority int    // Rule priority (higher values are evaluated first)
	Match    string // OVS-style match criteria (e.g., "in_port=1,dl_type=0x0800,nw_proto=6,tp_dst=80")
	Actions  string // OVS-style actions (e.g., "allow", "drop", "output:2")
}

// NetworkBackend abstracts Open vSwitch operations to decouple virtual networking from compute management.
// It provides a unified API for managing software-defined network (SDN) components.
type NetworkBackend interface {
	// Bridge Management

	// CreateBridge establishes a new virtual switch (OVS bridge) with a specific VXLAN VNI.
	CreateBridge(ctx context.Context, name string, vxlanID int) error
	// DeleteBridge removes a virtual switch and all its associated ports.
	DeleteBridge(ctx context.Context, name string) error
	// ListBridges returns the names of all currently configured virtual switches.
	ListBridges(ctx context.Context) ([]string, error)

	// Port Management

	// AddPort attaches a virtual or physical network interface to a bridge.
	AddPort(ctx context.Context, bridge, portName string) error
	// DeletePort disconnects an interface from a bridge.
	DeletePort(ctx context.Context, bridge, portName string) error

	// VXLAN Tunnels (multi-node overlay networks)

	// CreateVXLANTunnel establishes a tunnel endpoint for cross-host communication.
	CreateVXLANTunnel(ctx context.Context, bridge string, vni int, remoteIP string) error
	// DeleteVXLANTunnel removes a cross-host tunnel endpoint.
	DeleteVXLANTunnel(ctx context.Context, bridge string, remoteIP string) error

	// Security Groups (implemented via OpenFlow rules)

	// AddFlowRule inserts a new traffic filtering or routing rule into the virtual switch.
	AddFlowRule(ctx context.Context, bridge string, rule FlowRule) error
	// DeleteFlowRule removes flow rules matching the specified criteria.
	DeleteFlowRule(ctx context.Context, bridge string, match string) error
	// ListFlowRules retrieves all active OpenFlow rules for a bridge.
	ListFlowRules(ctx context.Context, bridge string) ([]FlowRule, error)

	// Veth Pair Management (used to link instance namespaces to the bridge)

	// CreateVethPair creates a linked pair of virtual ethernet interfaces.
	CreateVethPair(ctx context.Context, hostEnd, containerEnd string) error
	// AttachVethToBridge links one end of a veth pair to an OVS bridge.
	AttachVethToBridge(ctx context.Context, bridge, vethEnd string) error
	// DeleteVethPair removes both ends of a virtual ethernet pair.
	DeleteVethPair(ctx context.Context, hostEnd string) error
	// SetVethIP assigns an IP address to a virtual ethernet interface.
	SetVethIP(ctx context.Context, vethEnd, ip, cidr string) error

	// Health & Type

	// Ping verifies the connectivity and responsiveness of the networking service.
	Ping(ctx context.Context) error
	// Type returns the identifier of the networking implementation (e.g., "ovs").
	Type() string
}
