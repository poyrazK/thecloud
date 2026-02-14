# Implementation Plan: CloudLogs Service

This plan outlines the steps to implement the CloudLogs service, providing persistent logging for all platform resources.

## Phase 1: Domain & Ports
- [x] Create `internal/core/domain/log.go` with `LogEntry` and `LogQuery` types.
- [x] Create `internal/core/ports/log.go` defining `LogRepository` and `LogService`.

## Phase 2: Database Layer
- [x] Create SQL migration for `log_entries` table:
    - `id` (UUID)
    - `tenant_id` (UUID)
    - `resource_id` (string)
    - `resource_type` (string)
    - `level` (string: INFO, WARN, ERROR)
    - `message` (text)
    - `timestamp` (TIMESTAMPTZ)
    - `trace_id` (string, optional)
- [x] Implement `internal/repositories/postgres/log.go`.

## Phase 3: Core Service
- [x] Implement `internal/core/services/cloudlogs.go`.
- [x] Implement a `LogWorker` that performs periodic maintenance (retention).
- [x] Update `InstanceService` to trigger log ingestion on termination.

## Phase 4: API & Handlers
- [x] Create `internal/handlers/log_handler.go`.
- [x] Add routes:
    - `GET /logs`: Search and filter logs.
    - `GET /logs/:resource_id`: Get logs for a specific resource.
- [x] Wire dependencies in `internal/api/setup/dependencies.go`.

## Phase 5: SDK & CLI
- [x] Add `pkg/sdk/logs.go` with search and get methods.
- [x] Add `cmd/cloud/logs.go` with `search` and `show` commands.

## Phase 6: Verification
- [ ] Unit tests for `LogService`.
- [ ] Integration tests for `LogRepository`.
- [ ] Smoke test: Launch instance -> generate logs -> terminate -> verify logs still exist in CloudLogs.
