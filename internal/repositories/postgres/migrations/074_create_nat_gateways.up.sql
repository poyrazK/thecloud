-- +goose Up
CREATE TABLE IF NOT EXISTS nat_gateways (
    id UUID PRIMARY KEY,
    vpc_id UUID NOT NULL REFERENCES vpcs(id) ON DELETE CASCADE,
    subnet_id UUID NOT NULL REFERENCES subnets(id),
    elastic_ip_id UUID NOT NULL REFERENCES elastic_ips(id),
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    private_ip INET NOT NULL,
    arn VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nat_vpc ON nat_gateways(vpc_id);
CREATE INDEX IF NOT EXISTS idx_nat_subnet ON nat_gateways(subnet_id);
CREATE INDEX IF NOT EXISTS idx_nat_eip ON nat_gateways(elastic_ip_id);