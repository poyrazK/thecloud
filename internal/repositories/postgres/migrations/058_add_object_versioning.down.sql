-- +goose Up
-- (Handled in 058_add_object_versioning.up.sql)

-- +goose Down
ALTER TABLE object_versions DROP TABLE IF EXISTS object_versions;
ALTER TABLE objects DROP COLUMN IF EXISTS current_version_id;
ALTER TABLE objects DROP COLUMN IF EXISTS is_versioned;

-- Restore unique constraint if it was removed (it shouldn't be if we just drop columns, but for completeness)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = current_schema() AND constraint_name = 'objects_bucket_key_key') THEN
        ALTER TABLE objects ADD CONSTRAINT objects_bucket_key_key UNIQUE (bucket_id, key);
    END IF;
END $$;
