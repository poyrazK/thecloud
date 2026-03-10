-- +goose Down
DROP INDEX IF EXISTS idx_events_tenant;
DROP INDEX IF EXISTS idx_audit_logs_tenant;

ALTER TABLE events DROP COLUMN IF NOT EXISTS tenant_id;
ALTER TABLE audit_logs DROP COLUMN IF NOT EXISTS tenant_id;
