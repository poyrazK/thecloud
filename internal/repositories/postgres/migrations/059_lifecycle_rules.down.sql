-- +goose Up
-- Handled in .up.sql

-- +goose Down
DROP TABLE IF EXISTS lifecycle_rules;
