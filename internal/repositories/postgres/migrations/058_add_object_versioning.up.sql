-- +goose Up
-- Add versioning columns to objects
ALTER TABLE objects ADD COLUMN version_id VARCHAR(64);
ALTER TABLE objects ADD COLUMN is_latest BOOLEAN DEFAULT TRUE;

-- Existing objects are considered the "null" version
UPDATE objects SET version_id = 'null' WHERE version_id IS NULL;

-- Remove old unique constraint and add new one that includes version_id
-- Note: Constraint name might vary, but objects_bucket_key_key is standard for (bucket, key)
ALTER TABLE objects DROP CONSTRAINT IF EXISTS objects_bucket_key_key;
ALTER TABLE objects ADD CONSTRAINT objects_bucket_key_version_unique UNIQUE (bucket, key, version_id);

-- Performance index for latest version lookups
CREATE INDEX IF NOT EXISTS idx_objects_latest ON objects(bucket, key) WHERE is_latest = TRUE;

