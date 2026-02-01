-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_owner ON tenants(owner_id);

-- Create tenant_members junction table
CREATE TABLE IF NOT EXISTS tenant_members (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_user ON tenant_members(user_id);

-- Create tenant_quotas table
CREATE TABLE IF NOT EXISTS tenant_quotas (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    max_instances INT NOT NULL DEFAULT 10,
    max_vpcs INT NOT NULL DEFAULT 3,
    max_storage_gb INT NOT NULL DEFAULT 100,
    max_memory_gb INT NOT NULL DEFAULT 32,
    max_vcpus INT NOT NULL DEFAULT 16
);

-- Add default_tenant_id to users table
ALTER TABLE users ADD COLUMN default_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;

-- Add default_tenant_id to api_keys table
ALTER TABLE api_keys ADD COLUMN default_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
