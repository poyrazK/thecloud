-- +goose Up
CREATE TABLE IF NOT EXISTS buckets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_public BOOLEAN DEFAULT FALSE,
    versioning_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add foreign key to objects table if it doesn't exist
-- We use TEXT/VARCHAR for bucket name in objects table, so we can link it
-- ALTER TABLE objects ADD CONSTRAINT fk_objects_bucket FOREIGN KEY (bucket) REFERENCES buckets(name) ON DELETE RESTRICT;
-- Actually, let's just make sure the table exists for now to satisfy the repo queries.
