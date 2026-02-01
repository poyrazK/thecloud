-- +goose Up
ALTER TABLE buckets 
    ADD COLUMN encryption_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN encryption_key_id VARCHAR(64);

CREATE TABLE IF NOT EXISTS encryption_keys (
    id VARCHAR(64) PRIMARY KEY,
    bucket_name VARCHAR(255) NOT NULL,
    encrypted_key BYTEA NOT NULL,
    algorithm VARCHAR(32) DEFAULT 'AES-256-GCM',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(bucket_name)
);

