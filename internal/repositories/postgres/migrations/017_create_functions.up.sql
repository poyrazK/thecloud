-- +goose Up

CREATE TABLE IF NOT EXISTS functions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    runtime VARCHAR(50) NOT NULL,
    handler VARCHAR(255) NOT NULL,
    code_path TEXT NOT NULL,
    timeout_seconds INT DEFAULT 30,
    memory_mb INT DEFAULT 128,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS invocations (
    id UUID PRIMARY KEY,
    function_id UUID NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    duration_ms INT,
    status_code INT,
    logs TEXT
);

CREATE INDEX IF NOT EXISTS idx_invocations_function ON invocations(function_id);
