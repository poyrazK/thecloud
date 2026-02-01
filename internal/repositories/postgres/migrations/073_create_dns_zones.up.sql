CREATE TABLE IF NOT EXISTS dns_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vpc_id UUID NOT NULL REFERENCES vpcs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    default_ttl INTEGER NOT NULL DEFAULT 300,
    powerdns_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id, name),
    UNIQUE(vpc_id)  -- One zone per VPC
);

CREATE INDEX IF NOT EXISTS idx_dns_zones_vpc ON dns_zones(vpc_id);
CREATE INDEX IF NOT EXISTS idx_dns_zones_tenant ON dns_zones(tenant_id);
