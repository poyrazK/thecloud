-- +goose Up
ALTER TABLE databases ADD COLUMN pooling_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE databases ADD COLUMN pooling_port INT;
ALTER TABLE databases ADD COLUMN pooler_container_id VARCHAR(255);
