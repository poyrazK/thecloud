-- +goose Up
-- FunctionSchedules: Scheduled Function Invocations
CREATE TABLE IF NOT EXISTS function_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    payload BYTEA,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    claimed_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_function_schedules_user_id ON function_schedules(user_id);
CREATE INDEX IF NOT EXISTS idx_function_schedules_status_next_run ON function_schedules(status, next_run_at) WHERE status = 'ACTIVE';
CREATE INDEX IF NOT EXISTS idx_function_schedules_claimed_until ON function_schedules(claimed_until) WHERE claimed_until IS NOT NULL;

CREATE TABLE IF NOT EXISTS function_schedule_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES function_schedules(id) ON DELETE CASCADE,
    invocation_id UUID NOT NULL REFERENCES invocations(id),
    status VARCHAR(50) NOT NULL,
    status_code INT NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_function_schedule_runs_schedule_id ON function_schedule_runs(schedule_id);