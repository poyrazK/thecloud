-- +goose Up
CREATE TABLE IF NOT EXISTS route_tables (
    id UUID PRIMARY KEY,
    vpc_id UUID NOT NULL REFERENCES vpcs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_main BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(vpc_id, name)
);

CREATE INDEX IF NOT EXISTS idx_route_tables_vpc ON route_tables(vpc_id);

CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY,
    route_table_id UUID NOT NULL REFERENCES route_tables(id) ON DELETE CASCADE,
    destination_cidr CIDR NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    target_id UUID,
    target_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routes_table ON routes(route_table_id);
CREATE INDEX IF NOT EXISTS idx_routes_destination ON routes(destination_cidr);

CREATE TABLE IF NOT EXISTS route_table_associations (
    id UUID PRIMARY KEY,
    route_table_id UUID NOT NULL REFERENCES route_tables(id) ON DELETE CASCADE,
    subnet_id UUID NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(route_table_id, subnet_id)
);

CREATE INDEX IF NOT EXISTS idx_rt_assoc_subnet ON route_table_associations(subnet_id);