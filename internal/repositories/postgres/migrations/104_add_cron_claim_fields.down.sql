-- +goose Down
DROP INDEX IF EXISTS idx_cron_jobs_claimed_until;
ALTER TABLE cron_jobs DROP COLUMN IF EXISTS claimed_until;
