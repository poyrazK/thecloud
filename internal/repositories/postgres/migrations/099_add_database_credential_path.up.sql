-- +goose Up
ALTER TABLE databases ADD COLUMN credential_path VARCHAR(255);

-- +goose Down
ALTER TABLE databases DROP COLUMN credential_path;
