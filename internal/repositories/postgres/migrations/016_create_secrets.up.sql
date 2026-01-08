-- +goose Up

CREATE TABLE secrets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    encrypted_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    CONSTRAINT secrets_name_user_key UNIQUE (name, user_id)
);

CREATE INDEX idx_secrets_user ON secrets(user_id);
