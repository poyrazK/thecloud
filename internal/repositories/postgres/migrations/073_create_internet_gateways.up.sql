-- +goose Up
CREATE TABLE IF NOT EXISTS internet_gateways (
    id UUID PRIMARY KEY,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'detached',
    arn VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_igw_vpc ON internet_gateways(vpc_id);
CREATE INDEX IF NOT EXISTS idx_igw_tenant ON internet_gateways(tenant_id);