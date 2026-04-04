-- +goose Up
ALTER TABLE databases ADD COLUMN IF NOT EXISTS parameters JSONB DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE databases DROP COLUMN IF EXISTS parameters;
