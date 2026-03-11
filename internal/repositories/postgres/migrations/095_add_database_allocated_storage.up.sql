-- +goose Up
ALTER TABLE databases ADD COLUMN IF NOT EXISTS allocated_storage INT NOT NULL DEFAULT 10;

-- +goose Down
ALTER TABLE databases DROP COLUMN IF EXISTS allocated_storage;
