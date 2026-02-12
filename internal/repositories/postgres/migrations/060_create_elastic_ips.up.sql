CREATE TABLE IF NOT EXISTS elastic_ips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL,
    public_ip VARCHAR(45) NOT NULL UNIQUE,
    instance_id UUID REFERENCES instances(id) ON DELETE SET NULL,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'allocated',
    arn VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optimization indexes
CREATE INDEX idx_elastic_ips_tenant_id ON elastic_ips(tenant_id);
CREATE INDEX idx_elastic_ips_instance_id ON elastic_ips(instance_id);

-- Business rule: An instance can have at most one Elastic IP
CREATE UNIQUE INDEX idx_elastic_ips_instance_unique 
    ON elastic_ips(instance_id) 
    WHERE instance_id IS NOT NULL;
