# ADR 025: VPC Routing Architecture

## Status
Accepted

## Date
2026-04-24

## Context

The platform needed to implement full VPC routing capabilities to enable private subnet internet access, VPC peering connectivity, and centralized traffic management. The initial implementation of VPC Peering directly manipulated OVS flow rules instead of using a route table abstraction, creating tight coupling between business logic and network implementation details.

## Decision

We decided to implement a layered routing architecture with route table abstraction:

### 1. Route Table Domain Model

Route tables are first-class VPC resources that contain ordered routes:
```go
type RouteTable struct {
    ID     uuid.UUID
    VPCID  uuid.UUID
    Name   string
    IsMain bool  // One main RT per VPC, auto-created
    Routes []Route
}

type Route struct {
    ID              uuid.UUID
    RouteTableID    uuid.UUID
    DestinationCIDR string
    TargetType      RouteTargetType  // local, igw, nat, peering, instance
    TargetID        *uuid.UUID        // References IGW, NAT, Peering, Instance
    TargetName      string            // Human-readable name
}
```

### 2. Route Table Abstraction for VPC Peering

Instead of direct OVS flow manipulation, VPCPeeringService now adds/removes routes via the route table repository:

```go
// addPeeringFlows creates routes in main route tables
reqRT, _ := s.rtRepo.GetMainByVPC(ctx, requesterVPC.ID)
s.rtRepo.AddRoute(ctx, reqRT.ID, &Route{
    DestinationCIDR: accepterVPC.CIDRBlock,
    TargetType:      RouteTargetPeering,
    TargetID:        &peeringID,
})
```

Similarly `removePeeringFlows` uses `s.rtRepo.RemoveRoute`.

**Why:** This decouples VPC peering from OVS implementation. The same peering routes can be programmed via different backends (OVS, Linux bridge, cloud hypervisor).

### 3. Auto-Creation of Main Route Table

When a VPC is created, its main route table is auto-created with a local route:
```go
mainRT := &RouteTable{
    VPCID:  vpc.ID,
    Name:   "main",
    IsMain: true,
    Routes: []Route{{
        DestinationCIDR: vpc.CIDRBlock,
        TargetType:      RouteTargetLocal,
    }},
}
s.routeTableRepo.Create(ctx, mainRT)
```

### 4. NAT Gateway via OVS Adapter

NAT gateways use the `NetworkBackend` interface to set up iptables SNAT:
```go
// NetworkBackend interface
SetupNATForSubnet(ctx context.Context, bridge, natVethEnd, subnetCIDR, egressIP string) error
RemoveNATForSubnet(ctx context.Context, bridge, natVethEnd, subnetCIDR string) error
```

The OVS adapter implements this using veth pairs and iptables rules.

### 5. Service Layer Wiring

Services are injected via dependency injection in `dependencies.go`:
```go
type Services struct {
    VpcService         *services.VpcService
    VPCPeeringService  *services.VPCPeeringService
    RouteTableService  *services.RouteTableService
    InternetGateway    *services.InternetGatewayService
    NATGateway         *services.NATGatewayService
    // ...
}
```

## Consequences

### Positive
- **Abstraction**: Route tables provide a stable API for routing decisions. OVS flow changes are an implementation detail.
- **Consistency**: All VPC routing (peering, IGW, NAT) now uses the same route table model.
- **Testability**: Services can be unit tested with mock route table repositories.
- **Maintainability**: Adding new routing targets (e.g., VPN gateway) only requires implementing a new `TargetType`, not rewriting peering logic.

### Negative
- **Complexity**: Additional layer of indirection. Routes must be kept in sync with actual OVS flows.
- **Consistency Risk**: If OVS flows fail but route table records succeed, state becomes inconsistent. Mitigated by immediate rollback on flow programming failure.

### Neutral
- Main route table auto-creation means VPC creation now has a dependency on route table persistence.
- Route table operations add latency to peering accept/delete operations.

## Alternatives Considered

### Alternative 1: Direct OVS Flow Management in PeeringService
**Why rejected:** Tied peering lifecycle to OVS implementation. Changes to network topology required modifying business logic. Harder to test without actual OVS.

### Alternative 2: Separate Routing Service with OVS Synchronization
**Why rejected:** Over-engineered. A separate sync process introduces eventual consistency issues. The route table abstraction inside the same service provides consistency without added complexity.

### Alternative 3: Delegating All Routing to OVS Adapter
**Why rejected:** Would require the adapter to understand VPC resources (peering IDs, EIP associations). Adapter should be thin - it receives high-level commands like "setup NAT for subnet" rather than low-level flow rules.

## Implementation Notes

- Route target types: `local`, `igw`, `nat`, `peering`, `instance`
- Only main route table cannot be deleted
- Subnet association is one-to-one: a subnet can only belong to one route table
- NAT gateway creation requires EIP to be in `allocated` (not `associated`) state