# ADR 013: Compute Resilience and Metadata Enhancements

## Status
Accepted

## Context
As the platform scales, users need better ways to organize and filter their compute instances. Additionally, instance reliability for standalone (non-ASG) instances was previously manual, and Docker-based simulations lacked common cloud features like SSH key injection.

We needed to address:
1.  **Metadata Management**: Providing a flexible way to tag and categorize instances.
2.  **Automated Recovery**: Ensuring standalone instances in `ERROR` state are automatically recovered.
3.  **Feature Parity**: Simulating Cloud-Init behavior in Docker containers to match Libvirt/KVM capabilities.

## Decision
We have implemented three major enhancements to the Compute layer:

### 1. Unified Metadata & Labels API
We added `Metadata` (key-value) and `Labels` (indexed key-value) fields to the `Instance` domain entity.
- **Storage**: Persisted as `JSONB` in PostgreSQL for efficient querying and flexibility.
- **Interface**: Exposed via `POST /instances` (launch) and `PUT /instances/:id/metadata` (update).
- **Organization**: Metadata is for arbitrary context, while Labels are intended for system-level filtering (e.g., tiering).

### 2. Self-Healing Background Worker
A new `HealingWorker` was introduced to monitor instance health.
- **Logic**: Periodically lists all instances and identifies those in `StatusError`.
- **Action**: Orchestrates a `Restart` (Stop -> Start) to clear transient failures.
- **Safety**: Implements a safety delay to avoid race conditions with inflight provisioning.

### 3. Docker Cloud-Init (SSH) Simulation
To support SSH key management in Docker-based simulations:
- **Injection**: The Docker adapter now parses Cloud-Init `ssh_authorized_keys` from user-data.
- **Execution**: Automatically creates `.ssh` directories and `authorized_keys` files with appropriate permissions (0600) using `Docker Exec`.
- **Compatibility**: Allows testing of SSH-reliant workflows without requiring a full Libvirt environment.

## Consequences

### Positive
- Improved organization and service discovery potential via labels.
- Reduced manual intervention for failed instances.
- Better parity between development (Docker) and production (Libvirt) environments.

### Negative
- Increased complexity in the `InstanceRepository` and `InstanceService`.
- Background healing may hide underlying infrastructure issues if not monitored via audit logs.
