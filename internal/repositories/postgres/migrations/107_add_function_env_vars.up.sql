-- +goose Up

ALTER TABLE functions ADD COLUMN env_vars JSONB DEFAULT '{}';

-- +goose Down

ALTER TABLE functions DROP COLUMN IF EXISTS env_vars;
