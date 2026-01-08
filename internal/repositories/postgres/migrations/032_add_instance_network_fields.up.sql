-- +goose Up
ALTER TABLE instances ADD COLUMN subnet_id UUID REFERENCES subnets(id);
ALTER TABLE instances ADD COLUMN private_ip INET;
ALTER TABLE instances ADD COLUMN ovs_port VARCHAR(64);

CREATE INDEX idx_instances_subnet ON instances(subnet_id);
