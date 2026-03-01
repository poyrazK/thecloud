-- +goose Down
ALTER TABLE databases DROP COLUMN parameters;
