-- +goose Up

CREATE TABLE objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    arn VARCHAR(512) NOT NULL UNIQUE,
    bucket VARCHAR(255) NOT NULL,
    key VARCHAR(512) NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (bucket, key)
);
CREATE INDEX idx_objects_bucket ON objects(bucket);
