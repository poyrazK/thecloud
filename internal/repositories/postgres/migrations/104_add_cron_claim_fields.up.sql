-- +goose Up
-- Add claim tracking for distributed cron job execution
ALTER TABLE cron_jobs ADD COLUMN IF NOT EXISTS claimed_until TIMESTAMPTZ;

-- Index to find stale claims efficiently
CREATE INDEX IF NOT EXISTS idx_cron_jobs_claimed_until ON cron_jobs(claimed_until)
    WHERE claimed_until IS NOT NULL;
