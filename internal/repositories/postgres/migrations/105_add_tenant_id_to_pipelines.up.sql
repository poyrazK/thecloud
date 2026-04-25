-- +goose Up

-- Add tenant_id to pipelines table
ALTER TABLE pipelines ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_pipelines_tenant ON pipelines(tenant_id);

-- Backfill tenant_id for existing pipelines from owner's default_tenant_id
UPDATE pipelines p
SET tenant_id = u.default_tenant_id
FROM users u
WHERE p.user_id = u.id AND p.tenant_id IS NULL;

-- Add tenant_id to builds table (inherits from pipeline's tenant)
ALTER TABLE builds ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_builds_tenant ON builds(tenant_id);

-- Backfill tenant_id for existing builds from their pipeline's tenant_id
UPDATE builds b
SET tenant_id = p.tenant_id
FROM pipelines p
WHERE b.pipeline_id = p.id AND b.tenant_id IS NULL;
