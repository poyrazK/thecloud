-- +goose Up
-- CloudCron: Scheduled Tasks Service
CREATE TABLE cron_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    target_url TEXT NOT NULL,
    target_method VARCHAR(10) NOT NULL DEFAULT 'POST',
    target_payload TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_cron_jobs_user_id ON cron_jobs(user_id);
CREATE INDEX idx_cron_jobs_status_next_run ON cron_jobs(status, next_run_at) WHERE status = 'ACTIVE';

CREATE TABLE cron_job_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES cron_jobs(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    status_code INT,
    response TEXT,
    duration_ms BIGINT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cron_job_runs_job_id ON cron_job_runs(job_id);
