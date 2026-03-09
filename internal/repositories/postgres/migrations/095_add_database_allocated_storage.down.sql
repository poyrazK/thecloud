-- +goose Down
ALTER TABLE databases DROP COLUMN allocated_storage;