-- +goose Up

CREATE TABLE security_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vpc_id UUID NOT NULL REFERENCES vpcs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    arn VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(vpc_id, name)
);

CREATE TABLE security_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES security_groups(id) ON DELETE CASCADE,
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('ingress', 'egress')),
    protocol VARCHAR(10) NOT NULL,
    port_min INT,
    port_max INT,
    cidr VARCHAR(18) NOT NULL,
    priority INT DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE instance_security_groups (
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES security_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (instance_id, group_id)
);

CREATE INDEX idx_security_rules_group_id ON security_rules(group_id);
CREATE INDEX idx_instance_security_groups_instance ON instance_security_groups(instance_id);
