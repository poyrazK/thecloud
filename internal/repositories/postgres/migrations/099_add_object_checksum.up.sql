-- +goose Up
ALTER TABLE objects ADD COLUMN IF NOT EXISTS checksum VARCHAR(64);
ALTER TABLE objects ADD CONSTRAINT ck_objects_checksum_valid CHECK (checksum IS NULL OR checksum ~ '^[0-9a-f]{64}$');
