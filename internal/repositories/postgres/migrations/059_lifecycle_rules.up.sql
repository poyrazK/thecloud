-- +goose Up
CREATE TABLE IF NOT EXISTS lifecycle_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bucket_name VARCHAR(255) NOT NULL REFERENCES buckets(name) ON DELETE CASCADE,
    prefix VARCHAR(512) DEFAULT '',
    expiration_days INT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(bucket_name, prefix)
);

CREATE INDEX IF NOT EXISTS idx_lifecycle_rules_enabled ON lifecycle_rules(enabled) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_lifecycle_rules_bucket ON lifecycle_rules(bucket_name);

