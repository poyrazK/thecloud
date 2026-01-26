-- +goose Up
-- Handled in .up.sql

-- +goose Down
DROP INDEX IF EXISTS idx_objects_latest;
ALTER TABLE objects DROP CONSTRAINT IF EXISTS objects_bucket_key_version_unique;
ALTER TABLE objects ADD CONSTRAINT objects_bucket_key_key UNIQUE (bucket, key);
ALTER TABLE objects DROP COLUMN IF EXISTS version_id;
ALTER TABLE objects DROP COLUMN IF EXISTS is_latest;
