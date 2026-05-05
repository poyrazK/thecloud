-- +goose Up

ALTER TABLE databases DROP COLUMN credential_version;

-- +goose Down

ALTER TABLE databases ADD COLUMN credential_version INT DEFAULT 1 NOT NULL;
