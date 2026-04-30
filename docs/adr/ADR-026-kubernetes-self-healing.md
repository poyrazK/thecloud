# ADR-026: Kubernetes Cluster Self-Healing

## Status
Accepted

## Context

The `ClusterReconciler` worker periodically checks cluster health, but the `Repair()` method it called was a complete stub (`return nil`). Clusters with a broken API server or unhealthy nodes would never recover automatically — the reconciler detected problems but couldn't fix them.

The repair path was also used by the user-facing `RepairCluster` API, which had the same issue: it spawned an async goroutine that called the stub, did nothing, and returned success.

## Decision

We implemented a real `Repair()` method in `KubeadmProvisioner` that diagnoses what broke and applies targeted remediation.

### Health Detection

`GetHealth()` checks:
- API server reachability via `kubectl get nodes`
- Node readiness count vs total

Results are persisted to the cluster record: `IsHealthy`, `UnhealthySince`, `FailureReason`.

### Repair Strategies

| Condition | Action |
|-----------|--------|
| API server unreachable | Iterate all control plane IPs; restart kubelet on each node sequentially |
| Nodes not ready | Re-apply Calico CNI; restart kube-proxy daemonset; restart kubelet on non-ready nodes |

### Status Lifecycle

```
Running → (health check fails) → Repairing → (success) → Running
                                           → (failure) → Failed
```

`ClusterStatusRepairing` is now set at the start of repair (both auto and manual) to prevent concurrent repair attempts.

### Reconciler Backoff

The reconciler skips repair for a cluster if:
- It was successfully repaired within the last 5 minutes (`LastRepairSucceeded`)
- It has been unhealthy for less than 2 minutes (`UnhealthySince`) — transient tolerance

After a failed repair, the cluster is marked `Failed` and subsequent reconciliation cycles skip it until status returns to `Running`.

## Consequences

### Positive
- Unhealthy clusters automatically recover without user intervention
- Health state is now visible via the API (`is_healthy`, `unhealthy_since`, `failure_reason`, `repair_attempts`)
- Concurrent repair attempts are prevented via status guard
- Reconciler no longer hammers failing clusters

### Negative
- Repair operations that restart kubelet may cause brief pod disruption on the affected node
- HA control plane recovery depends on at least one control plane node having a functional kubelet
- Calico version is hardcoded in the repair logic (`v3.25.0`) — should track cluster's installed version
