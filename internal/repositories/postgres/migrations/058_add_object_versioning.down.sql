-- +goose Up
-- (Handled in 058_add_object_versioning.up.sql)

-- +goose Down
-- Remove versioning columns and index from objects
DROP INDEX IF EXISTS idx_objects_latest;
ALTER TABLE objects DROP CONSTRAINT IF EXISTS objects_bucket_key_version_unique;
ALTER TABLE objects DROP COLUMN IF EXISTS version_id;
ALTER TABLE objects DROP COLUMN IF EXISTS is_latest;

-- Restore original unique constraint on (bucket, key)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = current_schema() AND constraint_name = 'objects_bucket_key_key') THEN
        ALTER TABLE objects ADD CONSTRAINT objects_bucket_key_key UNIQUE (bucket, key);
    END IF;
END $$;
