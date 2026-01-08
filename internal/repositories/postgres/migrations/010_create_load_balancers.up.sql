-- +goose Up

CREATE TABLE load_balancers (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(64) UNIQUE,
    name VARCHAR(255) NOT NULL,
    vpc_id UUID REFERENCES vpcs(id),
    port INT NOT NULL,
    algorithm VARCHAR(50) DEFAULT 'round-robin',
    status VARCHAR(50) DEFAULT 'CREATING',
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_lb_vpc ON load_balancers(vpc_id);
CREATE INDEX idx_lb_status ON load_balancers(status);

CREATE TABLE lb_targets (
    id UUID PRIMARY KEY,
    lb_id UUID REFERENCES load_balancers(id) ON DELETE CASCADE,
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    port INT NOT NULL,
    weight INT DEFAULT 1,
    health VARCHAR(50) DEFAULT 'unknown',
    UNIQUE(lb_id, instance_id)
);

CREATE INDEX idx_lb_targets_lb ON lb_targets(lb_id);
