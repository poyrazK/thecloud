-- +goose Down
-- Remove tenant_id from the remaining resource tables

DROP INDEX IF EXISTS idx_api_keys_tenant;
ALTER TABLE api_keys DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_usage_records_tenant;
ALTER TABLE usage_records DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_databases_tenant;
ALTER TABLE databases DROP COLUMN IF EXISTS tenant_id;
