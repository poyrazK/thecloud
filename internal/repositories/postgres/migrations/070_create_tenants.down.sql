ALTER TABLE api_keys DROP COLUMN IF EXISTS default_tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS default_tenant_id;
DROP TABLE IF EXISTS tenant_quotas;
DROP TABLE IF EXISTS tenant_members;
DROP TABLE IF EXISTS tenants;
