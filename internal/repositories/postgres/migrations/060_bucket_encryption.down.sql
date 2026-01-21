-- +goose Up
-- Handled in up migration

-- +goose Down
DROP TABLE IF EXISTS encryption_keys;
ALTER TABLE buckets DROP COLUMN IF EXISTS encryption_key_id;
ALTER TABLE buckets DROP COLUMN IF EXISTS encryption_enabled;
