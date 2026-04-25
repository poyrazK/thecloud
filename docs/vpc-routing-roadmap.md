# Full VPC Routing - Stacked PR Roadmap

## Overview

Implementation divided into **5 sequential PRs** that stack onto a feature branch `feature/vpc-routing`, then merge to main as a complete feature.

```
main ──────────────────────────────────────────────────────────
         │
         └─ feature/vpc-routing ── PR#1 ── PR#2 ── PR#3 ── PR#4 ── PR#5 ── merge to main
```

**Workflow:**
1. Create `feature/vpc-routing` branch from main
2. Open PR #1 against `feature/vpc-routing`
3. After PR #1 merges, rebase PR #2 onto updated feature branch
4. Repeat until all PRs merged
5. Final PR: merge `feature/vpc-routing` into main

---

## PR #1: Domain Models & Ports

**Branch:** `feature/vpc-routing`
**Target:** `feature/vpc-routing`

### Goal
Establish the foundation - domain structs and interfaces only. No implementation logic.

### Files to Create

| File | Purpose |
|------|---------|
| `internal/core/domain/route_table.go` | RouteTable, Route, RouteTableAssociation, RouteTargetType |
| `internal/core/domain/internet_gateway.go` | InternetGateway, IGWStatus |
| `internal/core/domain/nat_gateway.go` | NATGateway, NATGatewayStatus |
| `internal/core/ports/route_table.go` | RouteTableRepository, RouteTableService interfaces |
| `internal/core/ports/internet_gateway.go` | IGWRepository, InternetGatewayService interfaces |
| `internal/core/ports/nat_gateway.go` | NATGatewayRepository, NATGatewayService interfaces |

### Validation
```bash
go build ./internal/core/domain/...
go build ./internal/core/ports/...
```

### PR Description
```
feat(vpc-routing): Add domain models and ports for Route Tables, IGW, and NAT Gateway

- RouteTable with Routes and Subnet associations
- InternetGateway with attach/detach lifecycle
- NATGateway with EIP dependency
- Repository and Service interfaces for each
```

---

## PR #2: Database Migrations & Repositories

**Branch:** `feature/vpc-routing`
**Target:** `feature/vpc-routing` (rebase after PR #1 merges)

### Goal
Add persistence layer - PostgreSQL tables and repository implementations.

### Files to Create

| File | Purpose |
|------|---------|
| `internal/repositories/postgres/migrations/XXX_create_route_tables.up.sql` | route_tables, routes, route_table_associations, internet_gateways, nat_gateways |
| `internal/repositories/postgres/migrations/XXX_add_subnet_routing_table.up.sql` | Add routing_table_id to subnets |
| `internal/repositories/postgres/route_table_repo.go` | RouteTableRepository implementation |
| `internal/repositories/postgres/igw_repo.go` | IGWRepository implementation |
| `internal/repositories/postgres/nat_gateway_repo.go` | NATGatewayRepository implementation |

### Files to Modify
- `internal/repositories/postgres/repository.go` - register new repositories

### Dependencies
- Requires PR #1 merged and `feature/vpc-routing` rebased

### Validation
```bash
go run github.com/pressly/goose/v3/cmd/goose up
go build ./internal/repositories/postgres/...
```

### PR Description
```
feat(vpc-routing): Add persistence layer for Route Tables, IGW, and NAT Gateway

- PostgreSQL migrations for 5 new tables
- Repository implementations with pgx
- Subnet now references a routing table
```

---

## PR #3: OVS Adapter Extensions

**Branch:** `feature/vpc-routing`
**Target:** `feature/vpc-routing` (rebase after PR #2 merges)

### Goal
Extend NetworkBackend with NAT support and implement in OVS adapter.

### Files to Modify

| File | Change |
|------|--------|
| `internal/core/ports/network.go` | Add SetupNATForSubnet, RemoveNATForSubnet to NetworkBackend |
| `internal/repositories/ovs/adapter.go` | Implement NAT methods with iptables SNAT |
| `internal/repositories/noop/network_adapter.go` | Noop stubs for NAT methods |

### NAT Implementation
- NAT Gateway runs as "bastion host" in public subnet
- Host-side veth connects to VPC bridge
- iptables SNAT for outbound traffic
- Return traffic auto de-SNATed via connection tracking

### Requirements
- `net.ipv4.ip_forward=1` must be enabled on the host system
  - This is a **system-wide setting** required for NAT functionality
  - Can be enabled with: `sysctl -w net.ipv4.ip_forward=1`
  - To persist across reboots, add to `/etc/sysctl.conf`: `net.ipv4.ip_forward=1`
  - Without this setting, NAT gateway traffic will not be forwarded

### Dependencies
- Requires PR #2 merged and `feature/vpc-routing` rebased

### Validation
```bash
go build ./internal/repositories/ovs/...
go build ./internal/repositories/noop/...
# Manual: iptables -t nat -L -n shows SNAT rules after setup
```

### PR Description
```
feat(vpc-routing): Extend NetworkBackend with NAT support

- Add SetupNATForSubnet/RemoveNATForSubnet to NetworkBackend interface
- OVS adapter implementation using iptables SNAT
- Noop adapter stubs for testing environments
```

---

## PR #4: Service Implementations

**Branch:** `feature/vpc-routing`
**Target:** `feature/vpc-routing` (rebase after PR #3 merges)

### Goal
Implement RouteTableService, InternetGatewayService, NATGatewayService. Integrate with existing services.

### Files to Create

| File | Purpose |
|------|---------|
| `internal/core/services/route_table.go` | RouteTableService: CRUD + route management + subnet associations |
| `internal/core/services/internet_gateway.go` | InternetGatewayService: create/attach/detach lifecycle |
| `internal/core/services/nat_gateway.go` | NATGatewayService: create with EIP, SNAT setup |

### Files to Modify

| File | Change |
|------|--------|
| `internal/core/services/vpc.go` | CreateVPC auto-creates main route table with local route |
| `internal/core/services/vpc_peering.go` | AcceptPeering adds routes, DeletePeering removes routes |
| `internal/api/setup/infrastructure.go` | Wire up new services |

### Key Integrations

**VpcService (CreateVPC):**
```go
mainRT := &domain.RouteTable{
    ID:     uuid.New(),
    VPCID:  vpc.ID,
    Name:   "main",
    IsMain: true,
}
mainRT.Routes = append(mainRT.Routes, domain.Route{
    ID:              uuid.New(),
    RouteTableID:    mainRT.ID,
    DestinationCIDR: vpc.CIDRBlock,
    TargetType:      domain.RouteTargetLocal,
})
routeTableRepo.Create(ctx, mainRT)
```

**VPCPeeringService (AcceptPeering):**
```go
reqRT, _ := routeTableRepo.GetMainByVPC(ctx, reqVPC.ID)
accRT, _ := routeTableRepo.GetMainByVPC(ctx, accVPC.ID)
routeTableRepo.AddRoute(ctx, reqRT.ID, &domain.Route{
    DestinationCIDR: accVPC.CIDRBlock,
    TargetType:      domain.RouteTargetPeering,
    TargetID:        &peering.ID,
})
// ... add reverse route to accepter RT
```

### Dependencies
- Requires PR #3 merged and `feature/vpc-routing` rebased

### Validation
```bash
go build ./internal/core/services/...
go test ./internal/core/services/... -run "RouteTable|InternetGateway|NATGateway" -v
```

### PR Description
```
feat(vpc-routing): Implement Route Table, IGW, and NAT Gateway services

- RouteTableService: CRUD + route management + subnet associations
- InternetGatewayService: create/attach/detach lifecycle
- NATGatewayService: create with EIP, SNAT setup
- VpcService now auto-creates main route table on VPC creation
- VPCPeeringService now manages route table entries on accept/delete
```

---

## PR #5: CLI Commands

**Branch:** `feature/vpc-routing`
**Target:** `feature/vpc-routing` (rebase after PR #4 merges)

### Goal
Add `cloud` CLI commands for route tables, IGW, and NAT Gateway.

### Files to Create

| File | Purpose |
|------|---------|
| `cmd/cloud/route_table.go` | route-table list/create/rm/add-route/associate/disassociate |
| `cmd/cloud/igw.go` | igw create/attach/detach/list/rm |
| `cmd/cloud/nat_gateway.go` | nat-gateway create/list/rm |

### Files to Modify
- `cmd/cloud/main.go` - register new CLI commands

### CLI Design

```bash
# Route Tables
cloud route-table list [vpc-id]
cloud route-table create [vpc-id] [name]        # --main flag for main RT
cloud route-table rm [rt-id]
cloud route-table add-route [rt-id] [dest-cidr] [target-type]  # --target-id
cloud route-table associate [rt-id] [subnet-id]
cloud route-table disassociate [rt-id] [subnet-id]

# Internet Gateway
cloud igw create
cloud igw attach [igw-id] [vpc-id]    # adds 0.0.0.0/0 route
cloud igw detach [igw-id]
cloud igw list
cloud igw rm [igw-id]                  # must be detached

# NAT Gateway
cloud nat-gateway create [subnet-id] [eip-id]
cloud nat-gateway list [vpc-id]
cloud nat-gateway rm [nat-id]
```

### Dependencies
- Requires PR #4 merged and `feature/vpc-routing` rebased

### Validation
```bash
go build ./cmd/cloud/...
cloud route-table --help
cloud igw --help
cloud nat-gateway --help
```

### PR Description
```
feat(vpc-routing): Add CLI commands for Route Tables, IGW, and NAT Gateway

- cloud route-table: list/create/rm/add-route/associate/disassociate
- cloud igw: create/attach/detach/list/rm
- cloud nat-gateway: create/list/rm
```

---

## Final Merge PR: Feature Branch → Main

**Branch:** `feature/vpc-routing`
**Target:** `main`

After all 5 PRs stacked and tested on feature branch:

```bash
# Create final merge PR
git checkout main
git pull
git merge feature/vpc-routing
git push origin main
```

Or via GitHub: PR from `feature/vpc-routing` → `main`

---

## Git Workflow Summary

```bash
# 1. Create feature branch
git checkout main
git pull origin main
git checkout -b feature/vpc-routing
git push -u origin feature/vpc-routing

# 2. PR #1: Domain + Ports
# ... implement ...
# Open PR targeting feature/vpc-routing
# After merge:
git fetch origin
git rebase origin/feature/vpc-routing

# 3. PR #2: Migrations + Repos (rebase then implement)
# ... implement ...
# Force push to PR branch, merge, repeat rebase

# 4-5. Repeat for remaining PRs

# 6. Final merge
git checkout main
git merge feature/vpc-routing
git push origin main

# Cleanup
git branch -d feature/vpc-routing
git push origin --delete feature/vpc-routing
```

---

## Testing Per PR

| PR | Unit Tests | Integration |
|----|-----------|-------------|
| #1 | None (types only) | None |
| #2 | Repository mocks | None |
| #3 | Adapter unit tests | Manual OVS verification |
| #4 | Service unit tests | VPC + RT creation flow |
| #5 | CLI argument parsing | End-to-end CLI smoke test |

---

## Rollback Plan

If `feature/vpc-routing` has issues at any point:

```bash
# Reset feature branch to main
git checkout feature/vpc-routing
git reset --hard origin/main
# Rebase remaining unmerged PRs if needed
```

Individual PR rollbacks:
1. PR #5: Revert CLI changes - services still work via API
2. PR #4: Revert service changes - database records remain but unused
3. PR #3: Revert OVS adapter - NAT methods become noop
4. PR #2: `goose rollback` + revert repo files
5. PR #1: Revert domain + port files

---

## Timeline Estimate

| PR | Complexity | Files | Notes |
|----|-----------|-------|-------|
| #1 | Low | 6 new | Pure types + interfaces |
| #2 | Medium | 6 new + 1 mod | Migrations + repo implementations |
| #3 | Medium | 3 modified | OVS/iptables integration |
| #4 | High | 3 new + 3 mod | Core business logic |
| #5 | Low | 3 new + 1 mod | CLI commands |