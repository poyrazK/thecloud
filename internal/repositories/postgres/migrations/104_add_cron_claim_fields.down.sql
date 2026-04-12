-- +goose Down
ALTER TABLE cron_jobs DROP COLUMN IF EXISTS claimed_until;
