-- +goose Up
-- handled in .up.sql

-- +goose Down
ALTER TABLE databases DROP COLUMN primary_id;
ALTER TABLE databases DROP COLUMN role;
