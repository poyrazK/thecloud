-- +goose Up
ALTER TABLE objects ADD COLUMN IF NOT EXISTS checksum VARCHAR(64);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_objects_checksum_valid') THEN
        ALTER TABLE objects ADD CONSTRAINT ck_objects_checksum_valid CHECK (checksum IS NULL OR checksum ~ '^[0-9a-f]{64}$');
    END IF;
END
$$;
