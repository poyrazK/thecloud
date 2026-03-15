-- +goose Up
-- Add tenant_id to remaining resource tables

-- 1. databases
ALTER TABLE databases ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_databases_tenant ON databases(tenant_id);

-- 2. usage_records
ALTER TABLE usage_records ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_usage_records_tenant ON usage_records(tenant_id);

-- 3. api_keys
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant ON api_keys(tenant_id);

-- Backfill existing data using default_tenant_id from users table
UPDATE databases d SET tenant_id = u.default_tenant_id FROM users u WHERE d.user_id = u.id AND d.tenant_id IS NULL;
UPDATE usage_records ur SET tenant_id = u.default_tenant_id FROM users u WHERE ur.user_id = u.id AND ur.tenant_id IS NULL;
UPDATE api_keys ak SET tenant_id = u.default_tenant_id FROM users u WHERE ak.user_id = u.id AND ak.tenant_id IS NULL;
