# PRD: CloudLogs Service

## Problem Statement
Currently, resource logs (instances, functions) are ephemeral and streamed directly from the compute backend. If an instance is terminated or a function execution finishes, the logs are lost. Users have no way to search historical logs, filter by time, or perform cross-resource log analysis, which is critical for production debugging and compliance.

## User Stories
- **As a Developer**, I want my instance logs to be persisted even after termination, so that I can debug post-mortem failures.
- **As an Operator**, I want to search logs across all my resources in a VPC using keywords, so that I can identify distributed system issues quickly.
- **As a Security Admin**, I want to retain logs for 30 days, so that I can comply with our organization's auditing policies.

## Acceptance Criteria
- [x] Implement a `CloudLogs` service following the hexagonal architecture.
- [x] Support log ingestion from `InstanceService` and `FunctionService`.
- [x] Provide a persistent storage backend (PostgreSQL initially, with a port for Loki).
- [x] Implement API endpoints for searching and filtering logs by `ResourceID`, `TenantID`, and `TimeRange`.
- [x] Add log rotation/retention policy (e.g., auto-delete logs older than X days).
- [x] Integrate TraceID into log entries for correlation with Jaeger.

## UX / CLI Design
```bash
# View historical logs
cloud logs tail <resource-id> --since 1h

# Search logs
cloud logs search "error" --resource-type instance --vpc <vpc-id>

# Configure retention
cloud logs set-retention --days 30
```

## Out of Scope
- Real-time log alerting (handled by CloudNotify bridge in a later phase).
- Log export to external S3 buckets (v2 feature).
