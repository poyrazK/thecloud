-- +goose Up

CREATE TABLE caches (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL DEFAULT 'redis',
    version VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'CREATING',
    vpc_id UUID REFERENCES vpcs(id),
    container_id VARCHAR(255),
    port INTEGER,
    password VARCHAR(255) NOT NULL,
    memory_mb INTEGER NOT NULL DEFAULT 128,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_caches_user_id ON caches(user_id);
