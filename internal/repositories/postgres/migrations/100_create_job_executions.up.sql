-- +goose Up
CREATE TABLE IF NOT EXISTS job_executions (
    job_key     TEXT        PRIMARY KEY,
    status      TEXT        NOT NULL DEFAULT 'running',   -- running | completed | failed
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    result      TEXT,
    -- Allow stale locks to be reclaimed: if a worker crashes while status='running',
    -- another worker can take over after started_at + timeout has elapsed.
    -- The timeout is enforced in application code, not in the schema.
    CONSTRAINT job_executions_status_check CHECK (status IN ('running', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_job_executions_status ON job_executions (status) WHERE status = 'running';
