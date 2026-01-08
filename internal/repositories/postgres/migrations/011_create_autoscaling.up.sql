-- +goose Up

-- Migration: 011_create_autoscaling.up.sql

CREATE TABLE scaling_groups (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(64) UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    vpc_id UUID REFERENCES vpcs(id),
    load_balancer_id UUID REFERENCES load_balancers(id) ON DELETE SET NULL,
    image VARCHAR(255) NOT NULL,
    ports VARCHAR(255),
    min_instances INT NOT NULL DEFAULT 1 CHECK (min_instances >= 0),
    max_instances INT NOT NULL DEFAULT 5 CHECK (max_instances <= 20),
    desired_count INT NOT NULL DEFAULT 1,
    current_count INT NOT NULL DEFAULT 0,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (min_instances <= max_instances),
    CHECK (desired_count >= min_instances AND desired_count <= max_instances)
);

CREATE TABLE scaling_policies (
    id UUID PRIMARY KEY,
    scaling_group_id UUID NOT NULL REFERENCES scaling_groups(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(50) NOT NULL DEFAULT 'cpu' CHECK (metric_type IN ('cpu', 'memory')),
    target_value DECIMAL(5,2) NOT NULL CHECK (target_value > 0 AND target_value <= 100),
    scale_out_step INT DEFAULT 1 CHECK (scale_out_step > 0),
    scale_in_step INT DEFAULT 1 CHECK (scale_in_step > 0),
    cooldown_sec INT DEFAULT 300 CHECK (cooldown_sec >= 60),
    last_scaled_at TIMESTAMPTZ,
    UNIQUE(scaling_group_id, name)
);

CREATE TABLE scaling_group_instances (
    scaling_group_id UUID NOT NULL REFERENCES scaling_groups(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY(scaling_group_id, instance_id)
);

-- Indexes
CREATE INDEX idx_sg_vpc ON scaling_groups(vpc_id);
CREATE INDEX idx_sg_status ON scaling_groups(status);
CREATE INDEX idx_sp_group ON scaling_policies(scaling_group_id);
CREATE INDEX idx_sgi_instance ON scaling_group_instances(instance_id);
CREATE INDEX idx_sgi_group_instance ON scaling_group_instances(scaling_group_id, instance_id);
