# ADR-025: Instance Resize/Scale

**Status**: Accepted
**Date**: 2026-04-24
**Deciders**: Platform Team

---

## Context

Instances launched on The Cloud are assigned an instance type (e.g., `basic-2`, `standard-8`) that determines CPU and memory allocation. Users running workloads that outgrow their current instance type need a way to scale up (or scale down for cost savings) without terminating and re-launching.

Existing lifecycle operations (Start, Stop, Terminate) were already implemented. Resize was the missing piece.

---

## Decision

We implemented `POST /instances/:id/resize` with an `instance_type` payload that changes the CPU and memory allocation of a running (or stopped) instance.

### Backend Strategy: Cold Resize for Libvirt

Libvirt does not support live CPU/memory hot-plug for the compute backends we target (KVM/QEMU). Therefore resize requires a **cold resize** cycle with a snapshot-before-resize safety net:

1. **Detect running state** via `DomainGetState`
2. **Stop the domain** (`DomainDestroy`) if currently running
3. **Fetch current domain XML** (`DomainGetXMLDesc`)
4. **Create a pre-resize snapshot** — QCOW2 volume is tar'd to `/tmp/snapshot-<id>-<name>.tar.gz`
5. **Patch the XML** using regex replacement of `<memory>`, `<currentMemory>`, and `<vcpu>` elements
6. **Undefine and redefine** the domain with new resources (`DomainUndefine`, `DomainDefineXML`)
7. **Restart the domain** (`DomainCreate`)
8. **On success**: delete the pre-resize snapshot

**Rollback on failure:** If any step fails after `DomainUndefine` (i.e. the domain has been undefined but the new domain hasn't been started), the system restores the volume from the snapshot archive and redefines the domain with the original XML via `DomainDefineXML`. This ensures the instance can be recovered to its pre-resize state even if the resize operation fails mid-way.

This is the same approach used by other Libvirt-based cloud platforms (OpenStack, oVirt) where live resize is not available.

### Docker: Warm Resize

Docker supports in-place container resource updates via `ContainerUpdate`, so Docker backend uses a direct update without restart. The same API is used but the implementation differs by backend.

### Quota Enforcement

The service calculates a **delta** between old and new instance types:

- If `deltaCPU > 0` or `deltaMem > 0` (upsize): calls `tenantSvc.CheckQuota` before proceeding
- If downsize (`delta < 0`): quota check is skipped (releasing resources back to the pool)
- If same size (`delta == 0`): no quota interaction

After a successful resize, usage counters are updated with the delta (`IncrementUsage` for upsize, `DecrementUsage` for downsize). Failures in usage updates are logged but not propagated — a future background reconciliation worker could correct drift.

### Error Handling

- Instance not found → `404 NotFound`
- Current or target instance type invalid → `400 InvalidInput`
- Quota exceeded → `403 Forbidden`
- Compute backend failure → `500 Internal` with metrics instrumentation (`resize_failure`)

---

## Consequences

### Positive
- Users can scale instance resources without destruction/recreation
- Quota enforcement prevents over-provisioning beyond tenant limits
- Multi-backend support (Docker warm, Libvirt cold) via unified interface
- Audit logging on every resize operation

### Negative
- Libvirt resize causes instance downtime (cold migration)
- Quota usage drift is possible if `IncrementUsage`/`DecrementUsage` calls fail silently
- Regex-based XML patching is fragile if domain XML format changes
- If snapshot creation fails, resize proceeds without a rollback safety net (logged as `WARN`)
- Rollback itself can fail if `RestoreSnapshot` or `DomainDefineXML` fails during recovery

### Neutral
- E2E tests require running server with Docker — skipped in unit/CI runs
- Handler validation (UUID parse, binding) runs before service call

---

## Alternatives Considered

### Alternative 1: Live Resize via Libvirt Live Migration
**Why rejected:** Live resize (hot-plug CPU/memory) requires QEMU guest agent support and is not reliably available across all VM images. Cold resize is the safe default for our target workloads.

### Alternative 2: Create New Instance and Migrate Data
**Why rejected:** Would require data copy steps, DNS/IP reconfiguration, and load balancer target updates. Far more complex than in-place resize.

### Alternative 3: Only Allow Upsize
**Why rejected:** Users with seasonal workloads legitimately need to downsize for cost savings. Full bidirectional resize is more useful.