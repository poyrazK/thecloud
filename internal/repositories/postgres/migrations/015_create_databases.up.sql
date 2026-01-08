-- +goose Up

CREATE TABLE databases (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    status VARCHAR(50) DEFAULT 'CREATING',
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    container_id VARCHAR(255),
    port INT,
    username VARCHAR(255),
    password VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_databases_user_id ON databases(user_id);
CREATE UNIQUE INDEX idx_databases_name_user ON databases(name, user_id);
