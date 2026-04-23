# ADR 025: Function Schedules (Cron Triggers for Serverless)

## Status
Accepted

## Context
The platform's serverless functions (`CloudFunctions`) support direct invocation via API but lack time-based triggering. Users cannot schedule function invocations on cron expressions. This is a fundamental gap for common serverless patterns like periodic data processing, nightly batch jobs, or maintenance tasks.

Existing infrastructure includes a `CronWorker` that polls PostgreSQL with `FOR UPDATE SKIP LOCKED` for distributed-safe job claiming. This pattern should be reused.

## Decision

We implement a dedicated `FunctionSchedule` system that parallels the existing `CronJob` infrastructure but targets function invocation instead of generic HTTP calls.

### 1. Domain Model

Two new domain types:

```go
type FunctionSchedule struct {
    ID           uuid.UUID
    UserID       uuid.UUID
    TenantID     uuid.UUID
    FunctionID   uuid.UUID  // Function to invoke
    Name         string
    Schedule     string     // Cron expression
    Payload      []byte      // Invocation payload
    Status       FunctionScheduleStatus  // ACTIVE, PAUSED, DELETED
    LastRunAt    *time.Time
    NextRunAt    *time.Time
    ClaimedUntil *time.Time // Distributed claiming timeout
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type FunctionScheduleRun struct {
    ID           uuid.UUID
    ScheduleID   uuid.UUID
    InvocationID uuid.UUID  // Links to Function.Invocation
    Status       string     // SUCCESS, FAILED
    StatusCode   int
    DurationMs   int64
    ErrorMessage string
    StartedAt    time.Time
}
```

### 2. Repository Layer

`FunctionScheduleRepository` mirrors `CronRepository`:

- `ClaimNextSchedulesToRun()` uses `FOR UPDATE SKIP LOCKED` with a 10-second poll interval and 5-minute claim timeout
- `CompleteScheduleRun()` atomically inserts the run record and advances `next_run_at`
- `ReapStaleClaims()` resets orphaned claims

### 3. Service Layer

`FunctionScheduleService` handles:
- **Create**: Validates cron expression (via `robfig/cron`), verifies function existence and user access, computes `next_run_at`
- **List/Get/Delete**: Standard CRUD with RBAC enforcement
- **Pause/Resume**: Updates status and recalculates `next_run_at` on resume

### 4. Worker

`FunctionScheduleWorker` reuses the `CronWorker` pattern:
- 10-second ticker polls for due schedules
- 1-minute ticker reaps stale claims
- Invokes functions via `FunctionService.InvokeFunction(ctx, functionID, payload, true)` (async mode)
- Status is captured from the returned `Invocation` record

### 5. RBAC Permissions

Four new permissions, following the existing `PermissionFunction*` pattern:

```go
PermissionFunctionScheduleCreate = "function_schedule:create"
PermissionFunctionScheduleRead   = "function_schedule:read"
PermissionFunctionScheduleDelete = "function_schedule:delete"
PermissionFunctionScheduleUpdate = "function_schedule:update"
```

### 6. API Routes

| Method | Path | Permission |
|--------|------|------------|
| `POST` | `/function-schedules` | `function_schedule:create` |
| `GET` | `/function-schedules` | `function_schedule:read` |
| `GET` | `/function-schedules/:id` | `function_schedule:read` |
| `DELETE` | `/function-schedules/:id` | `function_schedule:delete` |
| `POST` | `/function-schedules/:id/pause` | `function_schedule:update` |
| `POST` | `/function-schedules/:id/resume` | `function_schedule:update` |
| `GET` | `/function-schedules/:id/runs` | `function_schedule:read` |

### 7. Database Schema

Two new tables with migration number `107`:

```sql
CREATE TABLE function_schedules (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES functions(id),
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    payload BYTEA,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    claimed_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    UNIQUE(user_id, name)
);

CREATE TABLE function_schedule_runs (
    id UUID PRIMARY KEY,
    schedule_id UUID REFERENCES function_schedules(id),
    invocation_id UUID REFERENCES invocations(id),
    status VARCHAR(50),
    status_code INT,
    duration_ms BIGINT,
    error_message TEXT,
    started_at TIMESTAMPTZ
);
```

## Consequences

- **Positive**: Enables time-driven serverless patterns; distributed-safe by design; reuses proven cron claiming infrastructure
- **Positive**: Audit trail via `FunctionScheduleRun` records linked to `Invocation`
- **Neutral**: Separate tables from `cron_jobs` — avoids mixing concerns but duplicates infrastructure
- **Future**: Consider unifying `CronJob` and `FunctionSchedule` under a generic trigger abstraction (event-driven, webhook triggers) in later iterations
